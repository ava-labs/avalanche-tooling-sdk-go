// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package contract

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/rpc"
)

// GetSelector returns evm selector code for the given signature.
// EVM maps function and error signatures into 4-byte selectors.
// Works for both function signatures (e.g., "transfer(address,uint256)")
// and error signatures (e.g., "InvalidInitialization()").
func GetSelector(signature string) string {
	return "0x" + hex.EncodeToString(crypto.Keccak256([]byte(signature))[:4])
}

// GetRevertDataFromRPCError extracts the full revert data from an RPC error.
// Returns the complete revert data as a hex string if the error contains revert information.
// Returns an error if the RPC error doesn't contain revert data or the data is malformed.
func GetRevertDataFromRPCError(err error) (string, error) {
	if err == nil {
		return "", fmt.Errorf("no error provided")
	}

	// Unwrap error chain to find rpc.DataError
	var dataErr rpc.DataError
	if !errors.As(err, &dataErr) {
		return "", fmt.Errorf("error does not implement rpc.DataError interface")
	}

	// Get the error data
	errData := dataErr.ErrorData()
	if errData == nil {
		return "", fmt.Errorf("error data is nil")
	}

	// Error data can be a string (hex) or already bytes
	var revertDataHex string
	switch v := errData.(type) {
	case string:
		revertDataHex = strings.TrimPrefix(v, "0x")
	case []byte:
		revertDataHex = hex.EncodeToString(v)
	default:
		return "", fmt.Errorf("unexpected error data type: %T", errData)
	}

	return "0x" + revertDataHex, nil
}

// GetErrorSelectorFromRPCError extracts the error selector (first 4 bytes) from an RPC error.
// Returns the error selector as a hex string (e.g., "0x12345678") if the error contains revert data.
// Returns an error if the RPC error doesn't contain revert data or the data is malformed.
func GetErrorSelectorFromRPCError(err error) (string, error) {
	revertDataHex, err := GetRevertDataFromRPCError(err)
	if err != nil {
		return "", err
	}

	// Strip 0x prefix for length check
	hexData := strings.TrimPrefix(revertDataHex, "0x")

	// Need at least 8 hex characters (4 bytes) for the error selector
	if len(hexData) < 8 {
		return "", fmt.Errorf("revert data has less than 4 bytes")
	}

	// Extract error selector (first 8 hex characters = 4 bytes)
	errorSelector := "0x" + hexData[:8]
	return errorSelector, nil
}

// LookupErrorBySelector searches for a matching error in the signatureToError map
// by comparing error selectors. Returns the matched error if found, or an error indicating
// the selector was not found in the map.
func LookupErrorBySelector(
	errorSelector string,
	signatureToError map[string]error,
) (error, error) {
	// Match against known error signatures
	for errorSignature, mappedErr := range signatureToError {
		expectedSelector := GetSelector(errorSignature)
		if errorSelector == expectedSelector {
			return mappedErr, nil
		}
	}

	return nil, fmt.Errorf("unknown error selector: %s", errorSelector)
}

// ExtractErrorFromRPCError attempts to extract and map a Solidity error from an RPC error.
// It extracts the error selector from the RPC error and looks it up in the signatureToError map.
// If a match is found, returns the mapped error. Otherwise, returns the original error.
func ExtractErrorFromRPCError(rpcErr error, signatureToError map[string]error) error {
	// Try to extract error selector from RPC error
	errorSelector, selectorErr := GetErrorSelectorFromRPCError(rpcErr)
	if selectorErr != nil {
		return rpcErr
	}

	// Try to lookup the error in the error map
	mappedErr, lookupErr := LookupErrorBySelector(errorSelector, signatureToError)
	if lookupErr != nil {
		return rpcErr
	}

	return mappedErr
}

// ExtractAndEnrichRPCError attempts to map an RPC error, and if no mapping is found,
// enriches it with error selector and revert data.
// If a mapping is found, returns the mapped error.
// If no mapping is found, wraps the original error with selector and revert data if available.
func ExtractAndEnrichRPCError(rpcErr error, signatureToError map[string]error) error {
	// Try to extract error selector from RPC error
	errorSelector, selectorErr := GetErrorSelectorFromRPCError(rpcErr)
	if selectorErr != nil {
		return rpcErr
	}

	// Try to lookup the error in the error map
	mappedErr, lookupErr := LookupErrorBySelector(errorSelector, signatureToError)
	if lookupErr == nil {
		// Mapping found, return it
		return mappedErr
	}

	// No mapping found, enrich with selector and revert data
	revertData, revertErr := GetRevertDataFromRPCError(rpcErr)
	if revertErr == nil {
		return fmt.Errorf("%w (error selector: %s, revert data: %s)", rpcErr, errorSelector, revertData)
	}

	return fmt.Errorf("%w (error selector: %s)", rpcErr, errorSelector)
}
