// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blockchain

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

// CreateSubnetTx creates uncommitted CreateSubnetTx
// keychain in wallet will be used to build, sign and pay for the transaction
func (c *Subnet) CreateSubnetTx(wallet wallet.Wallet) (*tx.SignedTx, error) {
	if c.DeployInfo.ControlKeys == nil {
		return nil, fmt.Errorf("control keys are not provided")
	}
	if c.DeployInfo.Threshold == 0 {
		return nil, fmt.Errorf("threshold is not provided")
	}
	addrs := c.DeployInfo.ControlKeys
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
	ctx, cancel := utils.GetTimedContext(constants.SignatureTimeout)
	defer cancel()
	if err := wallet.P().Signer().Sign(ctx, &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return tx.New(&tx), nil
}

// CreateBlockchainTx creates uncommitted CreateChainTx
// keychain in wallet will be used to build, sign and pay for the transaction
func (c *Subnet) CreateBlockchainTx(wallet wallet.Wallet) (*tx.SignedTx, error) {
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
		return nil, fmt.Errorf("blockchain name is not provided")
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
	ctx, cancel := utils.GetTimedContext(constants.SignatureTimeout)
	defer cancel()
	if err := wallet.P().Signer().Sign(ctx, &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return tx.New(&tx), nil
}
