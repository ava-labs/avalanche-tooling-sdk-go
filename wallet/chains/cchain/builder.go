// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package cchain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// BuildTx builds C-Chain transactions
func BuildTx(wallet *primary.Wallet, params types.BuildTxParams) (types.BuildTxResult, error) {
	// TODO: Implement C-Chain transaction building
	return types.BuildTxResult{}, fmt.Errorf("C-Chain transactions not yet implemented")
}
