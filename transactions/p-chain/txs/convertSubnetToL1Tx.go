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

func (d *PublicDeployer) NewConvertSubnetToL1Tx(params ConvertSubnetToL1TxParams) (*multisig.Multisig, error) {
	options := d.getMultisigTxOptions(params.SubnetAuthKeys)
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
	tx := txs.Tx{Unsigned: unsignedTx}
	if err := params.Wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return multisig.New(&tx), nil
}
