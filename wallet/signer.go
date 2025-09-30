// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

// SignTx constructs a transaction for the specified operation
func SignTx(ctx context.Context, wallet *primary.Wallet, params SignTxParams) (tx.SignTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return tx.SignTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case "P-Chain":
		return signPChainTx(ctx, wallet, params)
	case "C-Chain":
		return tx.SignTxResult{}, fmt.Errorf("C-Chain signing not yet implemented")
	case "X-Chain":
		return tx.SignTxResult{}, fmt.Errorf("X-Chain signing not yet implemented")
	default:
		return tx.SignTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

func signPChainTx(ctx context.Context, wallet *primary.Wallet, params SignTxParams) (tx.SignTxResult, error) {
	// Get the P-Chain transaction from the BuildTxResult
	pChainTx, ok := params.BuildTxResult.GetTx().(*avagoTxs.Tx)
	if !ok {
		return tx.SignTxResult{}, fmt.Errorf("expected P-Chain transaction, got %T", params.BuildTxResult.GetTx())
	}

	if err := wallet.P().Signer().Sign(context.Background(), pChainTx); err != nil {
		return tx.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	return tx.SignTxResult{Tx: pChainTx}, nil
}
