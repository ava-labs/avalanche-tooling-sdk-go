// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	utilsSDK "github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	avagokeychain "github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	commonAvago "github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"time"
)

type Wallet struct {
	*primary.Wallet
	Keychain keychain.Keychain
	options  []common.Option
	config   primary.WalletConfig
}

func New(ctx context.Context, uri string, avaxKeychain avagokeychain.Keychain, config primary.WalletConfig) (Wallet, error) {
	wallet, err := primary.MakeWallet(
		ctx,
		uri,
		avaxKeychain,
		secp256k1fx.NewKeychain(),
		config,
	)
	return Wallet{
		Wallet: wallet,
		Keychain: keychain.Keychain{
			Keychain: avaxKeychain,
		},
		config: config,
	}, err
}

// SecureWalletIsChangeOwner ensures that a fee paying address (wallet's keychain) will receive
// the change UTXO and not a randomly selected auth key that may not be paying fees
func (w *Wallet) SecureWalletIsChangeOwner() {
	addrs := w.Addresses()
	changeAddr := addrs[0]
	// sets change to go to wallet addr (instead of any other subnet auth key)
	changeOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{changeAddr},
	}
	w.options = append(w.options, common.WithChangeOwner(changeOwner))
	w.Wallet = primary.NewWalletWithOptions(w.Wallet, w.options...)
}

// SetAuthKeys sets auth keys that will be used when signing txs, besides the wallet's Keychain fee
// paying ones
func (w *Wallet) SetAuthKeys(authKeys []ids.ShortID) {
	addrs := w.Addresses()
	addrsSet := set.Set[ids.ShortID]{}
	addrsSet.Add(addrs...)
	addrsSet.Add(authKeys...)
	w.options = append(w.options, common.WithCustomAddresses(addrsSet))
	w.Wallet = primary.NewWalletWithOptions(w.Wallet, w.options...)
}

func (w *Wallet) SetSubnetAuthMultisig(authKeys []ids.ShortID) {
	w.SecureWalletIsChangeOwner()
	w.SetAuthKeys(authKeys)
}

func (w *Wallet) Addresses() []ids.ShortID {
	return w.Keychain.Addresses().List()
}

func (w *Wallet) GetMultisigTxOptions(subnetAuthKeys []ids.ShortID) []common.Option {
	options := []common.Option{}
	walletAddrs := w.Keychain.Addresses().List()
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

func (w *Wallet) Commit(ms multisig.Multisig, waitForTxAcceptance bool) (ids.ID, error) {
	if ms.Undefined() {
		return ids.Empty, multisig.ErrUndefinedTx
	}
	isReady, err := ms.IsReadyToCommit()
	if err != nil {
		return ids.Empty, err
	}
	if !isReady {
		return ids.Empty, errors.New("tx is not fully signed so can't be committed")
	}
	tx, err := ms.GetWrappedPChainTx()
	if err != nil {
		return ids.Empty, err
	}
	const (
		repeats             = 3
		sleepBetweenRepeats = 2 * time.Second
	)
	var issueTxErr error
	if err != nil {
		return ids.Empty, err
	}
	for i := 0; i < repeats; i++ {
		ctx, cancel := utilsSDK.GetAPILargeContext()
		defer cancel()
		options := []commonAvago.Option{commonAvago.WithContext(ctx)}
		if !waitForTxAcceptance {
			options = append(options, commonAvago.WithAssumeDecided())
		}
		// TODO: split error checking and recovery between issuing and waiting for status
		issueTxErr = w.P().IssueTx(tx, options...)
		if issueTxErr == nil {
			break
		}
		if ctx.Err() != nil {
			issueTxErr = fmt.Errorf("timeout issuing/verifying tx with ID %s: %w", tx.ID(), issueTxErr)
		} else {
			issueTxErr = fmt.Errorf("error issuing tx with ID %s: %w", tx.ID(), issueTxErr)
		}
		time.Sleep(sleepBetweenRepeats)
	}
	if issueTxErr != nil {
		return ids.Empty, fmt.Errorf("issue tx error %w", issueTxErr)
	}
	return tx.ID(), issueTxErr
}
