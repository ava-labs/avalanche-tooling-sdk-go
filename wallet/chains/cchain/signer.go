// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package cchain

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

// SignTx signs C-Chain transactions
func SignTx(ctx context.Context, wallet *primary.Wallet, params types.SignTxParams) (types.SignTxResult, error) {
	// TODO: Implement C-Chain signing when C-Chain is implemented
	return types.SignTxResult{}, fmt.Errorf("C-Chain signing not yet implemented")
}
