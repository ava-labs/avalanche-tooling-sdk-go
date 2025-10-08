// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package pchain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// SendTx sends P-Chain transactions
func SendTx(wallet *primary.Wallet, params types.SendTxParams) (types.SendTxResult, error) {
	// Get the P-Chain transaction from the SignTxResult
	pChainTx, ok := params.SignTxResult.GetTx().(*avagoTxs.Tx)
	if !ok {
		return types.SendTxResult{}, fmt.Errorf("expected P-Chain transaction, got %T", params.SignTxResult.GetTx())
	}

	// Submit the signed transaction to the network
	if err := wallet.P().IssueTx(pChainTx); err != nil {
		return types.SendTxResult{}, fmt.Errorf("error sending tx: %w", err)
	}
	pChainSendResult := types.NewPChainSendTxResult(pChainTx)
	return types.SendTxResult{SendTxOutput: pChainSendResult}, nil
}
