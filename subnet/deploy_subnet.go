// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"avalanche-tooling-sdk-go/avalanche"
	"avalanche-tooling-sdk-go/utils"
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

// createSubnetTx creates uncommitted createSubnet transaction
func createSubnetTx(subnet *Subnet, wallet primary.Wallet) (*txs.Tx, error) {
	addrs, err := address.ParseToIDs(subnet.ControlKeys)
	if err != nil {
		return nil, fmt.Errorf("failure parsing control keys: %w", err)
	}
	owners := &secp256k1fx.OutputOwners{
		Addrs:     addrs,
		Threshold: subnet.Threshold,
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

// createBlockchainTx creates uncommitted createBlockchain transaction
func createBlockchainTx(subnet *Subnet, wallet primary.Wallet, network avalanche.Network, keyChain avalanche.Keychain) (*txs.Tx, error) {
	wallet, err := loadCacheWallet(network, keyChain, wallet, subnet.SubnetID, subnet.TransferSubnetOwnershipTxID)
	if err != nil {
		return nil, err
	}
	fxIDs := make([]ids.ID, 0)
	options := getMultisigTxOptions(keyChain.Keychain, subnet.SubnetAuthKeys)
	// create tx
	unsignedTx, err := wallet.P().Builder().NewCreateChainTx(
		subnet.SubnetID,
		subnet.Genesis,
		subnet.VMID,
		fxIDs,
		subnet.Name,
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

func loadCacheWallet(network avalanche.Network, keyChain avalanche.Keychain, wallet primary.Wallet, preloadTxs ...ids.ID) (primary.Wallet, error) {
	var err error
	if wallet == nil {
		wallet, err = loadWallet(network, keyChain, preloadTxs...)
	}
	return wallet, err
}

func loadWallet(network avalanche.Network, keyChain avalanche.Keychain, preloadTxs ...ids.ID) (primary.Wallet, error) {
	ctx := context.Background()
	// filter out ids.Empty txs
	filteredTxs := utils.Filter(preloadTxs, func(e ids.ID) bool { return e != ids.Empty })
	wallet, err := primary.MakeWallet(
		ctx,
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keyChain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(filteredTxs...),
		},
	)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
