// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package cchain

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// SendTx sends C-Chain transactions
func SendTx(ctx context.Context, wallet *primary.Wallet, params types.SendTxParams) (types.SendTxResult, error) {
	// TODO: Implement C-Chain sending when C-Chain is implemented
	return types.SendTxResult{}, fmt.Errorf("C-Chain sending not yet implemented")
}
