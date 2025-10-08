// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package xchain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// SendTx sends X-Chain transactions
func SendTx(wallet *primary.Wallet, params types.SendTxParams) (types.SendTxResult, error) {
	// TODO: Implement X-Chain sending when X-Chain is implemented
	return types.SendTxResult{}, fmt.Errorf("X-Chain sending not yet implemented")
}
