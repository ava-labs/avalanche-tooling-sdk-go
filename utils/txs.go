// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"github.com/ava-labs/avalanchego/wallet/chain/x/builder"
	"github.com/ava-labs/coreth/plugin/evm/atomic"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"

	avmtxs "github.com/ava-labs/avalanchego/vms/avm/txs"
	platformvmtxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// AutoDetectChain attempts to determine which chain a transaction belongs to by analyzing its bytes.
// It returns UndefinedAlias if it cannot determine the chain or if multiple chains are possible (codec collision).
func AutoDetectChain(txBytes []byte) constants.ChainAlias {
	matchCount := 0
	detectedChain := constants.UndefinedAlias

	// Try to unmarshal as P-chain transaction
	var unsignedPTx platformvmtxs.UnsignedTx
	_, err := platformvmtxs.Codec.Unmarshal(txBytes, &unsignedPTx)
	if err == nil {
		matchCount++
		detectedChain = constants.PChainAlias
	}

	// Try to unmarshal as X-chain transaction
	var unsignedXTx avmtxs.UnsignedTx
	_, err = builder.Parser.Codec().Unmarshal(txBytes, &unsignedXTx)
	if err == nil {
		matchCount++
		detectedChain = constants.XChainAlias
	}

	// Try to unmarshal as C-chain transaction
	var unsignedCTx atomic.UnsignedAtomicTx
	_, err = atomic.Codec.Unmarshal(txBytes, &unsignedCTx)
	if err == nil {
		matchCount++
		detectedChain = constants.CChainAlias
	}

	// If more than one chain successfully unmarshaled the bytes, we have a codec collision
	if matchCount > 1 {
		return constants.UndefinedAlias
	}

	return detectedChain
}

// GetNetworkID extracts the network ID from P-Chain, X-Chain, or C-Chain transaction bytes.
// Returns 0 if the transaction cannot be unmarshaled or network ID cannot be determined.
func GetNetworkID(txBytes []byte) uint32 {
	// Try P-Chain
	var unsignedPTx platformvmtxs.UnsignedTx
	_, err := platformvmtxs.Codec.Unmarshal(txBytes, &unsignedPTx)
	if err == nil {
		return GetPChainTxNetworkID(unsignedPTx)
	}

	// Try X-Chain
	var unsignedXTx avmtxs.UnsignedTx
	_, err = builder.Parser.Codec().Unmarshal(txBytes, &unsignedXTx)
	if err == nil {
		return GetXChainTxNetworkID(unsignedXTx)
	}

	// Try C-Chain
	var unsignedCTx atomic.UnsignedAtomicTx
	_, err = atomic.Codec.Unmarshal(txBytes, &unsignedCTx)
	if err == nil {
		return GetCChainTxNetworkID(unsignedCTx)
	}

	return 0
}

// GetPChainTxNetworkID extracts the network ID from a P-Chain unsigned transaction.
// Returns 0 if the network ID cannot be determined from the transaction type.
func GetPChainTxNetworkID(tx platformvmtxs.UnsignedTx) uint32 {
	switch t := tx.(type) {
	case *platformvmtxs.AddDelegatorTx:
		return t.NetworkID
	case *platformvmtxs.AddPermissionlessValidatorTx:
		return t.NetworkID
	case *platformvmtxs.AddPermissionlessDelegatorTx:
		return t.NetworkID
	case *platformvmtxs.AddSubnetValidatorTx:
		return t.NetworkID
	case *platformvmtxs.AddValidatorTx:
		return t.NetworkID
	case *platformvmtxs.BaseTx:
		return t.NetworkID
	case *platformvmtxs.ConvertSubnetToL1Tx:
		return t.NetworkID
	case *platformvmtxs.CreateChainTx:
		return t.NetworkID
	case *platformvmtxs.CreateSubnetTx:
		return t.NetworkID
	case *platformvmtxs.DisableL1ValidatorTx:
		return t.NetworkID
	case *platformvmtxs.ExportTx:
		return t.NetworkID
	case *platformvmtxs.ImportTx:
		return t.NetworkID
	case *platformvmtxs.IncreaseL1ValidatorBalanceTx:
		return t.NetworkID
	case *platformvmtxs.RegisterL1ValidatorTx:
		return t.NetworkID
	case *platformvmtxs.RemoveSubnetValidatorTx:
		return t.NetworkID
	case *platformvmtxs.SetL1ValidatorWeightTx:
		return t.NetworkID
	case *platformvmtxs.TransferSubnetOwnershipTx:
		return t.NetworkID
	case *platformvmtxs.TransformSubnetTx:
		return t.NetworkID
	default:
		return 0
	}
}

// GetXChainTxNetworkID extracts the network ID from an X-Chain unsigned transaction.
// Returns 0 if the network ID cannot be determined from the transaction type.
func GetXChainTxNetworkID(tx avmtxs.UnsignedTx) uint32 {
	switch t := tx.(type) {
	case *avmtxs.BaseTx:
		return t.NetworkID
	case *avmtxs.CreateAssetTx:
		return t.NetworkID
	case *avmtxs.OperationTx:
		return t.NetworkID
	case *avmtxs.ImportTx:
		return t.NetworkID
	case *avmtxs.ExportTx:
		return t.NetworkID
	default:
		return 0
	}
}

// GetCChainTxNetworkID extracts the network ID from a C-Chain unsigned atomic transaction.
// Returns 0 if the network ID cannot be determined from the transaction type.
func GetCChainTxNetworkID(tx atomic.UnsignedAtomicTx) uint32 {
	switch t := tx.(type) {
	case *atomic.UnsignedImportTx:
		return t.NetworkID
	case *atomic.UnsignedExportTx:
		return t.NetworkID
	default:
		return 0
	}
}
