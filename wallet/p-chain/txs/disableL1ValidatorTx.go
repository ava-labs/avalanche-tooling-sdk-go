// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// DisableL1ValidatorTxParams contains all parameters needed to create a DisableL1ValidatorTx
type DisableL1ValidatorTxParams struct {
	// ValidationID is Validation ID of the validator in L1.
	ValidationID ids.ID
	// Wallet is the wallet used to sign `ConvertSubnetToL1Tx`
	Wallet *wallet.Wallet
}

func NewDisableL1ValidatorTx(params DisableL1ValidatorTxParams) (*multisig.Multisig, error) {
	unsignedTx, err := params.Wallet.P().Builder().NewDisableL1ValidatorTx(
		params.ValidationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error building tx: %w", err)
	}
	tx := txs.Tx{Unsigned: unsignedTx}
	if err := params.Wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return multisig.New(&tx), nil
}
