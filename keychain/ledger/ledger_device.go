// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ledger

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/version"

	ledger "github.com/ava-labs/ledger-avalanche-go"
	bip32 "github.com/tyler-smith/go-bip32"
)

const (
	rootPath = "m/44'/9000'/0'" // BIP44: m / purpose' / coin_type' / account'
	// ledgerBufferLimit corresponds to FLASH_BUFFER_SIZE for Nano S.
	// Modern devices (Nano X, Nano S2, Stax, Flex) support up to 16384 bytes,
	// but we use the conservative limit for universal compatibility.
	//
	// Ref: https://github.com/ava-labs/ledger-avalanche/blob/main/app/src/common/tx.c
	ledgerBufferLimit = 8192
	// ledgerPathSize is the serialized size of a path suffix (e.g., "0/123") as produced
	// by SerializePathSuffix in the ledger library. Format: 1 byte (component count) +
	// 4 bytes (first component) + 4 bytes (second component) = 9 bytes total.
	//
	// Ref: https://github.com/ava-labs/ledger-avalanche-go/blob/main/common.go (SerializePathSuffix)
	ledgerPathSize    = 9
	maxRetries        = 5
	initialRetryDelay = 200 * time.Millisecond
)

// retryOnHIDAPIError executes a function up to maxRetries times if it encounters
// the specific "hidapi: unknown failure" error or APDU error 0x6987
func retryOnHIDAPIError(fn func() error) error {
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		// These errors indicate transient USB communication issues that often resolve on retry:
		// - "hidapi: unknown failure": USB communication error from the HIDAPI library
		// - APDU 0x6987: "Interrupted execution" - occurs when the device is busy or communication is disrupted
		if err.Error() == "hidapi: unknown failure" || err.Error() == "APDU Error Code from Ledger Device: 0x6987" {
			if attempt < maxRetries {
				// Calculate backoff delay: 200ms, 400ms, 600ms, 800ms
				// This linear backoff prevents excessive delay (at most 2s), while at the same
				// time it is quick enough in most cases. Also proved to successfully recover
				// in all tests.
				delay := time.Duration(attempt) * initialRetryDelay
				time.Sleep(delay)
				continue
			}
		}

		// If it's not a retryable error or we've exhausted retries, exit the loop
		break
	}
	return err
}

// Device is a wrapper around the low-level Ledger Device interface that
// provides Avalanche-specific access.
type Device struct {
	device *ledger.LedgerAvalanche
	epk    *bip32.Key
}

func New() (*Device, error) {
	var device *ledger.LedgerAvalanche
	err := retryOnHIDAPIError(func() error {
		var err error
		device, err = ledger.FindLedgerAvalancheApp()
		return err
	})
	return &Device{
		device: device,
	}, err
}

func addressPath(index uint32) string {
	return fmt.Sprintf("%s/0/%d", rootPath, index)
}

func (l *Device) PubKey(addressIndex uint32) (*secp256k1.PublicKey, error) {
	var resp *ledger.ResponseAddr
	err := retryOnHIDAPIError(func() error {
		var err error
		resp, err = l.device.GetPubKey(addressPath(addressIndex), false, "", "")
		return err
	})
	if err != nil {
		return nil, err
	}
	pubKey, err := secp256k1.ToPublicKey(resp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failure parsing public key from ledger: %w", err)
	}
	return pubKey, nil
}

func (l *Device) PubKeys(addressIndices []uint32) ([]*secp256k1.PublicKey, error) {
	if l.epk == nil {
		var pk []byte
		var chainCode []byte
		err := retryOnHIDAPIError(func() error {
			var err error
			pk, chainCode, err = l.device.GetExtPubKey(rootPath, false, "", "")
			return err
		})
		if err != nil {
			return nil, err
		}
		l.epk = &bip32.Key{
			Key:       pk,
			ChainCode: chainCode,
		}
	}
	// derivation path rootPath/0 (BIP44 change level, when set to 0, known as external chain)
	externalChain, err := l.epk.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	pubKeys := make([]*secp256k1.PublicKey, len(addressIndices))
	for i, addressIndex := range addressIndices {
		// derivation path rootPath/0/v (BIP44 address index level)
		address, err := externalChain.NewChildKey(addressIndex)
		if err != nil {
			return nil, err
		}
		pubKey, err := secp256k1.ToPublicKey(address.Key)
		if err != nil {
			return nil, fmt.Errorf("failure parsing public key for ledger child key %d: %w", addressIndex, err)
		}
		pubKeys[i] = pubKey
	}
	return pubKeys, nil
}

func convertToSigningPaths(input []uint32) []string {
	output := make([]string, len(input))
	for i, v := range input {
		output[i] = fmt.Sprintf("0/%d", v)
	}
	return output
}

func (l *Device) SignHash(hash []byte, addressIndices []uint32) ([][]byte, error) {
	strIndices := convertToSigningPaths(addressIndices)
	var response *ledger.ResponseSign
	err := retryOnHIDAPIError(func() error {
		var err error
		response, err = l.device.SignHash(rootPath, strIndices, hash)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("%w: unable to sign hash", err)
	}
	responses := make([][]byte, len(addressIndices))
	for i, index := range strIndices {
		sig, ok := response.Signature[index]
		if !ok {
			return nil, fmt.Errorf("missing signature %s", index)
		}
		responses[i] = sig
	}
	return responses, nil
}

func (l *Device) Sign(txBytes []byte, addressIndices []uint32) ([][]byte, error) {
	// We pass addressIndices both as signing paths and change paths.
	// The ledger library deduplicates them, so the buffer contains len(addressIndices) paths.
	// Buffer format: 1 byte (path count) + paths + transaction bytes
	//
	// Ref: https://github.com/ava-labs/ledger-avalanche-go/blob/main/common.go (ConcatMessageAndChangePath)
	bufferSize := 1 + len(addressIndices)*ledgerPathSize + len(txBytes)
	if bufferSize > ledgerBufferLimit {
		// When the tx that is being signed is too large for the ledger buffer,
		// we sign with hash instead.
		unsignedHash := hashing.ComputeHash256(txBytes)
		return l.SignHash(unsignedHash, addressIndices)
	}
	strIndices := convertToSigningPaths(addressIndices)
	var response *ledger.ResponseSign
	err := retryOnHIDAPIError(func() error {
		var err error
		response, err = l.device.Sign(rootPath, strIndices, txBytes, strIndices)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("%w: unable to sign transaction", err)
	}
	responses := make([][]byte, len(strIndices))
	for i, index := range strIndices {
		sig, ok := response.Signature[index]
		if !ok {
			return nil, fmt.Errorf("missing signature %s", index)
		}
		responses[i] = sig
	}
	return responses, nil
}

func (l *Device) Version() (*version.Semantic, error) {
	var resp *ledger.VersionInfo
	err := retryOnHIDAPIError(func() error {
		var err error
		resp, err = l.device.GetVersion()
		return err
	})
	if err != nil {
		return nil, err
	}
	return &version.Semantic{
		Major: int(resp.Major),
		Minor: int(resp.Minor),
		Patch: int(resp.Patch),
	}, nil
}

func (l *Device) Disconnect() error {
	return retryOnHIDAPIError(func() error {
		return l.device.Close()
	})
}
