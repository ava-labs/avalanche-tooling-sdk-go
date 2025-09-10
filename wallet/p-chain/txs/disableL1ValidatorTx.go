// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	"github.com/ava-labs/avalanchego/ids"
)

// DisableL1ValidatorTxParams contains all parameters needed to create a DisableL1ValidatorTx
type DisableL1ValidatorTxParams struct {
	// ValidationID is Validation ID of the validator in L1.
	ValidationID ids.ID
	// Wallet is the wallet used to sign `ConvertSubnetToL1Tx`
	Wallet *wallet.Wallet
}

// GetTxType returns the transaction type identifier
func (p DisableL1ValidatorTxParams) GetTxType() string {
	return "DisableL1ValidatorTx"
}

// Validate validates the parameters
func (p DisableL1ValidatorTxParams) Validate() error {
	if p.ValidationID == ids.Empty {
		return fmt.Errorf("ValidationID cannot be empty")
	}
	if p.Wallet == nil {
		return fmt.Errorf("Wallet cannot be nil")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p DisableL1ValidatorTxParams) GetChainType() string {
	return "P-Chain"
}

//func NewDisableL1ValidatorTx(params DisableL1ValidatorTxParams) (*tx.SignTxResult, error) {
//	unsignedTx, err := params.Wallet.P().Builder().NewDisableL1ValidatorTx(
//		params.ValidationID,
//	)
//	if err != nil {
//		return nil, fmt.Errorf("error building tx: %w", err)
//	}
//	tx := txs.Tx{Unsigned: unsignedTx}
//	if err := params.Wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
//		return nil, fmt.Errorf("error signing tx: %w", err)
//	}
//	return tx.New(&tx), nil
//}
