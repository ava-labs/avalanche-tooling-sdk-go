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

// SendTx submits a signed transaction to the Network
func SendTx(ctx context.Context, wallet *primary.Wallet, params SendTxParams) (tx.SendTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return tx.SendTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.SignTxResult.GetChainType(); chainType {
	case "P-Chain":
		return sendPChainTx(ctx, wallet, params)
	case "C-Chain":
		return sendCChainTx(ctx, wallet, params)
	case "X-Chain":
		return sendXChainTx(ctx, wallet, params)
	default:
		return tx.SendTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

func sendPChainTx(ctx context.Context, wallet *primary.Wallet, params SendTxParams) (tx.SendTxResult, error) {
	// Get the P-Chain transaction from the SignTxResult
	pChainTx, ok := params.SignTxResult.GetTx().(*avagoTxs.Tx)
	if !ok {
		return tx.SendTxResult{}, fmt.Errorf("expected P-Chain transaction, got %T", params.SignTxResult.GetTx())
	}

	// Submit the signed transaction to the network
	if err := wallet.P().IssueTx(pChainTx); err != nil {
		return tx.SendTxResult{}, fmt.Errorf("error sending tx: %w", err)
	}
	return *tx.NewPChainSendTxResult(pChainTx), nil
}

func sendCChainTx(ctx context.Context, wallet *primary.Wallet, params SendTxParams) (tx.SendTxResult, error) {
	// TODO: Implement C-Chain sending when C-Chain is implemented
	return tx.SendTxResult{}, fmt.Errorf("C-Chain sending not yet implemented")
}

func sendXChainTx(ctx context.Context, wallet *primary.Wallet, params SendTxParams) (tx.SendTxResult, error) {
	// TODO: Implement X-Chain sending when X-Chain is implemented
	return tx.SendTxResult{}, fmt.Errorf("X-Chain sending not yet implemented")
}
