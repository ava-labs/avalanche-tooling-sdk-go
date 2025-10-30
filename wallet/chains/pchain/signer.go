// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package pchain

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// SignTx signs P-Chain transactions
func SignTx(wallet *primary.Wallet, params types.SignTxParams) (types.SignTxResult, error) {
	// Get the P-Chain transaction from the BuildTxResult
	pChainTx, ok := params.BuildTxResult.GetTx().(*avagoTxs.Tx)
	if !ok {
		return types.SignTxResult{}, fmt.Errorf("expected P-Chain transaction, got %T", params.BuildTxResult.GetTx())
	}

	if err := wallet.P().Signer().Sign(context.Background(), pChainTx); err != nil {
		return types.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	pChainSignResult := types.NewPChainSignTxResult(pChainTx)
	return types.SignTxResult{SignTxOutput: pChainSignResult}, nil
}
