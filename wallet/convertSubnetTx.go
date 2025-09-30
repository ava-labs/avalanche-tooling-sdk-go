// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func (w *LocalWallet) BuildConvertSubnetToL1Tx(ctx context.Context, account account.Account, params *txs.ConvertSubnetToL1TxParams) (tx.BuildTxResult, error) {
	options := getMultisigTxOptions(account, params.SubnetAuthKeys)
	unsignedTx, err := w.P().Builder().NewConvertSubnetToL1Tx(
		params.SubnetID,
		params.ChainID,
		params.Address,
		params.Validators,
		options...,
	)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	return tx.BuildTxResult{Tx: &builtTx}, nil
}
