// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"avalanche-tooling-sdk-go/keychain"
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

type Wallet struct {
	primary.Wallet
	Keychain keychain.Keychain
	Options  []common.Option
}

func New(ctx context.Context, config *primary.WalletConfig) (Wallet, error) {
	wallet, err := primary.MakeWallet(
		ctx,
		config,
	)
	return Wallet{
		Wallet: wallet,
		Keychain: keychain.Keychain{
			Keychain: config.AVAXKeychain,
		},
	}, err
}

// SecureWalletIsChangeOwner ensures that a fee paying address (wallet's Keychain) will receive
// the change UTXO and not a randomly selected auth key that may not be paying fees
func (w *Wallet) SecureWalletIsChangeOwner() {
	addrs := w.Keychain.Addresses().List()
	changeAddr := addrs[0]
	// sets change to go to wallet addr (instead of any other subnet auth key)
	changeOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{changeAddr},
	}
	w.Options = append(w.Options, common.WithChangeOwner(changeOwner))
	w.Wallet = primary.NewWalletWithOptions(w.Wallet, w.Options...)
}

// SetAuthKeys sets auth keys that will be used when signing txs, besides the wallet's Keychain fee
// paying ones
func (w *Wallet) SetAuthKeys(authKeys []ids.ShortID) {
	addrs := w.Keychain.Addresses().List()
	addrsSet := set.Set[ids.ShortID]{}
	addrsSet.Add(addrs...)
	addrsSet.Add(authKeys...)
	w.Options = append(w.Options, common.WithCustomAddresses(addrsSet))
	w.Wallet = primary.NewWalletWithOptions(w.Wallet, w.Options...)
}

func (w *Wallet) SetSubnetAuthMultisig(authKeys []ids.ShortID) {
	w.SecureWalletIsChangeOwner()
	w.SetAuthKeys(authKeys)
}
