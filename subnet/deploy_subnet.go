// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"

	"avalanche-tooling-sdk-go/avalanche"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

// CreateSubnetTx creates uncommitted createSubnet transaction
func (c *Subnet) CreateSubnetTx(wallet primary.Wallet) (*txs.Tx, error) {
	if c.DeployInfo.ControlKeys == nil {
		return nil, fmt.Errorf("control keys are not provided")
	}
	if c.DeployInfo.SubnetAuthKeys == nil {
		return nil, fmt.Errorf("subnet authkeys are not provided")
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
	return &tx, nil
}

// CreateBlockchainTx creates uncommitted createBlockchain transaction
func (c *Subnet) CreateBlockchainTx(wallet primary.Wallet, keyChain avalanche.Keychain) (*txs.Tx, error) {
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
	fxIDs := make([]ids.ID, 0)
	options := getMultisigTxOptions(keyChain.Keychain, c.DeployInfo.SubnetAuthKeys)
	// create tx
	unsignedTx, err := wallet.P().Builder().NewCreateChainTx(
		c.SubnetID,
		c.Genesis,
		c.VMID,
		fxIDs,
		c.Name,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("error building tx: %w", err)
	}
	tx := txs.Tx{Unsigned: unsignedTx}
	// sign with current wallet
	if err := wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return &tx, nil
}

func getMultisigTxOptions(keychain keychain.Keychain, subnetAuthKeys []ids.ShortID) []common.Option {
	options := []common.Option{}
	walletAddrs := keychain.Addresses().List()
	changeAddr := walletAddrs[0]
	// addrs to use for signing
	customAddrsSet := set.Set[ids.ShortID]{}
	customAddrsSet.Add(walletAddrs...)
	customAddrsSet.Add(subnetAuthKeys...)
	options = append(options, common.WithCustomAddresses(customAddrsSet))
	// set change to go to wallet addr (instead of any other subnet auth key)
	changeOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{changeAddr},
	}
	options = append(options, common.WithChangeOwner(changeOwner))
	return options
}
