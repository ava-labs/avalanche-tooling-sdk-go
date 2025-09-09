// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// ConvertSubnetToL1TxParams contains all parameters needed to create a ConvertSubnetToL1Tx
type ConvertSubnetToL1TxParams struct {
	// SubnetAuthKeys are the keys used to sign `ConvertSubnetToL1Tx`
	SubnetAuthKeys []ids.ShortID
	// SubnetID is Subnet ID of the subnet to convert to an L1.
	SubnetID ids.ID
	// ChainID is Blockchain ID of the L1 where the validator manager contract is deployed.
	ChainID ids.ID
	// Address is address of the validator manager contract.
	Address []byte
	// Validators are the initial set of L1 validators after the conversion.
	Validators []*txs.ConvertSubnetToL1Validator
	// Wallet is the wallet used to sign `ConvertSubnetToL1Tx`
	Wallet *wallet.Wallet
}

// GetTxType returns the transaction type identifier
func (p ConvertSubnetToL1TxParams) GetTxType() string {
	return "ConvertSubnetToL1Tx"
}

// Validate validates the parameters
func (p ConvertSubnetToL1TxParams) Validate() error {
	if p.SubnetID == ids.Empty {
		return fmt.Errorf("SubnetID cannot be empty")
	}
	if p.ChainID == ids.Empty {
		return fmt.Errorf("ChainID cannot be empty")
	}
	if len(p.Address) == 0 {
		return fmt.Errorf("Address cannot be empty")
	}
	if len(p.Validators) == 0 {
		return fmt.Errorf("Validators cannot be empty")
	}
	if p.Wallet == nil {
		return fmt.Errorf("Wallet cannot be nil")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p ConvertSubnetToL1TxParams) GetChainType() string {
	return "P-Chain"
}

func NewConvertSubnetToL1Tx(params ConvertSubnetToL1TxParams) (*tx.SignedTx, error) {
	options := params.Wallet.GetMultisigTxOptions(params.SubnetAuthKeys)
	unsignedTx, err := params.Wallet.P().Builder().NewConvertSubnetToL1Tx(
		params.SubnetID,
		params.ChainID,
		params.Address,
		params.Validators,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := txs.Tx{Unsigned: unsignedTx}
	if err := params.Wallet.P().Signer().Sign(context.Background(), &builtTx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return &tx.SignedTx{Tx: &builtTx}, nil
}
