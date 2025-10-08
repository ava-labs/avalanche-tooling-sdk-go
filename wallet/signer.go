// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/chains/pchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// SignTx constructs a transaction for the specified operation
func SignTx(ctx context.Context, wallet *primary.Wallet, params types.SignTxParams) (types.SignTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SignTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case pchain.ChainType:
		result, err := pchain.SignTx(wallet, params)
		if err != nil {
			return types.SignTxResult{}, err
		}
		return types.SignTxResult{SignTxOutput: result.SignTxOutput}, nil
	default:
		return types.SignTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}
