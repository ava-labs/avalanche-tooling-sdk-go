// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/chains/pchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// SendTx submits a signed transaction to the Network
func SendTx(wallet *primary.Wallet, params types.SendTxParams) (types.SendTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SendTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.SignTxResult.GetChainType(); chainType {
	case pchain.ChainType:
		result, err := pchain.SendTx(wallet, params)
		if err != nil {
			return types.SendTxResult{}, err
		}
		return types.SendTxResult{SendTxOutput: result.SendTxOutput}, nil
	default:
		return types.SendTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}
