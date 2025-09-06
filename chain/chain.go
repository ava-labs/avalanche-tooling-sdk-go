// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/cchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/pchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/xchain"
)

// ChainType represents the type of Avalanche chain
type ChainType int

const (
	// Undefined represents an unknown or unrecognized chain type
	Undefined ChainType = iota
	// CChain represents the Contract Chain (C-Chain)
	CChain
	// XChain represents the Exchange Chain (X-Chain)
	XChain
	// PChain represents the Platform Chain (P-Chain)
	PChain
)

// String returns the string representation of the ChainType
func (ct ChainType) String() string {
	switch ct {
	case CChain:
		return "C"
	case XChain:
		return "X"
	case PChain:
		return "P"
	case Undefined:
		return "undefined"
	default:
		return "undefined"
	}
}

// DetectTxChainType analyzes transaction bytes and returns the detected chain type.
// It uses a safeguard mechanism to prevent cross-chain confusion:
// - If exactly one chain recognizes the bytes, returns that chain type
// - If zero or multiple chains recognize the bytes, returns Undefined
// This ensures safe chain detection without ambiguity.
func DetectTxChainType(txBytes []byte) ChainType {
	// Check all three chain types
	isCChain := cchain.IsCChainTx(txBytes)
	isXChain := xchain.IsXChainTx(txBytes)
	isPChain := pchain.IsPChainTx(txBytes)

	// Count how many chains recognize this transaction
	detectionCount := 0
	if isCChain {
		detectionCount++
	}
	if isXChain {
		detectionCount++
	}
	if isPChain {
		detectionCount++
	}

	// Only return a specific chain type if exactly one chain recognizes the bytes
	// This prevents cross-chain confusion and ambiguity
	if detectionCount == 1 {
		if isCChain {
			return CChain
		}
		if isXChain {
			return XChain
		}
		if isPChain {
			return PChain
		}
	}

	// If no chains recognize it or multiple chains recognize it, return Undefined
	return Undefined
}
