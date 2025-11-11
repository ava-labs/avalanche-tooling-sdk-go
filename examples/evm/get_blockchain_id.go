//go:build get_blockchain_id
// +build get_blockchain_id

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
)

const warpPrecompileAddress = "0x0200000000000000000000000000000000000005"

// This example demonstrates how to get the blockchain ID using ReadContract
//
// Environment variables:
//   - RPC_URL: The RPC URL to connect to (defaults to Fuji C-Chain if not set)
func main() {
	// Get RPC URL from environment variable, default to "C" for Fuji C-Chain
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "C"
	}

	// Create a local wallet with Fuji network
	w, err := local.NewLocalWallet(logging.NoLog{}, network.FujiNetwork())
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	// Set the EVM chain
	if err := w.SetChain(rpcURL); err != nil {
		fmt.Printf("Failed to set chain: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to: %s\n", w.Chain())

	// Call the Warp precompile's getBlockchainID method
	fmt.Println("\nReading blockchain ID from Warp precompile...")
	warpPrecompile := common.HexToAddress(warpPrecompileAddress)
	method := wallet.Method("getBlockchainID()->(bytes32)")
	result, err := w.ReadContract(warpPrecompile, method)
	if err != nil {
		fmt.Printf("Failed to read blockchain ID: %s\n", err)
		os.Exit(1)
	}

	blockchainIDBytes, err := wallet.ParseResult[[32]byte](result)
	if err != nil {
		fmt.Printf("Failed to parse result: %s\n", err)
		os.Exit(1)
	}
	blockchainID, err := ids.ToID(blockchainIDBytes[:])
	if err != nil {
		fmt.Printf("Failed to convert to ID: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Blockchain ID: %s\n", blockchainID)
	fmt.Println("\n✓ Successfully retrieved blockchain ID")
}
