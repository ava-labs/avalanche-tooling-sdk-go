// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"

	"avalanche-tooling-sdk-go/multisig"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

// CreateSubnetTx creates uncommitted createSubnet transaction
func (c *Subnet) CreateSubnetTx(wallet Wallet) (*multisig.Multisig, error) {
	addrs, err := address.ParseToIDs(c.ControlKeys)
	if err != nil {
		return nil, fmt.Errorf("failure parsing control keys: %w", err)
	}
	owners := &secp256k1fx.OutputOwners{
		Addrs:     addrs,
		Threshold: c.Threshold,
		Locktime:  0,
	}
	unsignedTx, err := wallet.P().Builder().NewCreateSubnetTx(
		owners,
	)
	if err != nil {
		return nil, fmt.Errorf("error building tx: %w", err)
	}
	tx := txs.Tx{Unsigned: unsignedTx}
	if err := wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return multisig.New(&tx), nil
}

// CreateBlockchainTx creates uncommitted createBlockchain transaction
func (c *Subnet) CreateBlockchainTx(wallet Wallet) (*multisig.Multisig, error) {
	wallet.SetSubnetAuthMultisig(c.SubnetAuthKeys)
	// create tx
	fxIDs := make([]ids.ID, 0)
	unsignedTx, err := wallet.P().Builder().NewCreateChainTx(
		c.SubnetID,
		c.Genesis,
		c.VMID,
		fxIDs,
		c.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("error building tx: %w", err)
	}
	tx := txs.Tx{Unsigned: unsignedTx}
	if err := wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return multisig.New(&tx), nil
}
