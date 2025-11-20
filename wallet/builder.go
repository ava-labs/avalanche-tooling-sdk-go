// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/chains/pchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// BuildTx constructs a transaction for the specified operation
func BuildTx(wallet *primary.Wallet, account account.Account, params types.BuildTxParams) (types.BuildTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.BuildTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case pchain.ChainType:
		result, err := pchain.BuildTx(wallet, account, params)
		if err != nil {
			return types.BuildTxResult{}, err
		}
		return types.BuildTxResult{BuildTxOutput: result.BuildTxOutput}, nil
	default:
		return types.BuildTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}
