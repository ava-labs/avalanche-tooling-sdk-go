// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"errors"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

var ErrNoSubnetAuthKeysInWallet = errors.New("auth wallet does not contain auth keys")

type PublicDeployer struct {
	kc      *keychain.Keychain
	network network.Network
	wallet  *primary.Wallet
}

func NewPublicDeployer(kc *keychain.Keychain, network network.Network) *PublicDeployer {
	return &PublicDeployer{
		kc:      kc,
		network: network,
	}
}

func (d *PublicDeployer) getMultisigTxOptions(subnetAuthKeys []ids.ShortID) []common.Option {
	options := []common.Option{}
	walletAddrs := d.kc.Addresses().List()
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
