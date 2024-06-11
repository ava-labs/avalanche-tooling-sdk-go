// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

// CreateSubnetTx creates uncommitted CreateSubnetTx
// TODO: add which key is paying -> wallet.keychain?
func (c *Subnet) CreateSubnetTx(wallet wallet.Wallet) (*multisig.Multisig, error) {
	if c.DeployInfo.ControlKeys == nil {
		return nil, fmt.Errorf("control keys are not provided")
	}
	if c.DeployInfo.Threshold == 0 {
		return nil, fmt.Errorf("threshold is not provided")
	}
	addrs, err := address.ParseToIDs(c.DeployInfo.ControlKeys)
	if err != nil {
		return nil, fmt.Errorf("failure parsing control keys: %w", err)
	}
	owners := &secp256k1fx.OutputOwners{
		Addrs:     addrs,
		Threshold: c.DeployInfo.Threshold,
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

// CreateBlockchainTx creates uncommitted CreateChainTx
func (c *Subnet) CreateBlockchainTx(wallet wallet.Wallet) (*multisig.Multisig, error) {
	if c.SubnetID == ids.Empty {
		return nil, fmt.Errorf("subnet ID is not provided")
	}
	if c.DeployInfo.SubnetAuthKeys == nil {
		return nil, fmt.Errorf("subnet authkeys are not provided")
	}
	if c.Genesis == nil {
		return nil, fmt.Errorf("threshold is not provided")
	}
	if c.VMID == ids.Empty {
		return nil, fmt.Errorf("vm ID is not provided")
	}
	if c.Name == "" {
		return nil, fmt.Errorf("subnet name is not provided")
	}
	wallet.SetSubnetAuthMultisig(c.DeployInfo.SubnetAuthKeys)

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
