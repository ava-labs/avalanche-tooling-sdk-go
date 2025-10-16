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
