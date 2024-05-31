// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

var (
	ErrNoRemainingAuthSignersInWallet = errors.New("wallet does not contain any remaining auth signer")
	ErrNotReadyToCommit               = errors.New("tx is not fully signed so can't be commited")
)

type Wallet struct {
	primary.Wallet
	keychain keychain.Keychain
	options  []common.Option
	config   *primary.WalletConfig
}

func New(ctx context.Context, config *primary.WalletConfig) (Wallet, error) {
	wallet, err := primary.MakeWallet(
		ctx,
		config,
	)
	return Wallet{
		Wallet: wallet,
		keychain: keychain.Keychain{
			Keychain: config.AVAXKeychain,
		},
		config: config,
	}, err
}

func (w *Wallet) ResetKeychain(ctx context.Context, kc *keychain.Keychain) error {
	w.config.AVAXKeychain = kc.Keychain
	wallet, err := primary.MakeWallet(
		ctx,
		w.config,
	)
	if err != nil {
		return err
	}
	w.Wallet = wallet
	w.keychain = *kc
	return nil
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

// SetAuthKeys sets auth keys that will be used when signing txs, besides the wallet's keychain fee
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
	return w.keychain.Addresses().List()
}

func (w *Wallet) Sign(
	ms multisig.Multisig,
	checkAuth bool,
	commitIfReady bool,
	waitForTxAcceptanceOnCommit bool,
) (bool, bool, ids.ID, error) {
	if ms.Undefined() {
		return false, false, ids.Empty, multisig.ErrUndefinedTx
	}
	if w.keychain.LedgerEnabled() {
		// let's see if there is we can find remaining subnet auth keys in the LedgerEnabled
		_, remaining, err := ms.GetRemainingAuthSigners()
		if err != nil {
			return false, false, ids.Empty, err
		}
		kc := w.keychain
		oldCount := len(kc.Addresses())
		err = kc.AddLedgerAddresses(remaining)
		if err != nil {
			return false, false, ids.Empty, err
		}
		if len(kc.Addresses()) != oldCount {
			ctx, cancel := utils.GetAPIContext()
			defer cancel()
			if err := w.ResetKeychain(ctx, &kc); err != nil {
				return false, false, ids.Empty, err
			}
		}
	}
	if checkAuth {
		remainingInWallet, err := w.GetRemainingAuthSignersInWallet(ms)
		if err != nil {
			return false, false, ids.Empty, fmt.Errorf("error signing tx: %w", err)
		}
		if len(remainingInWallet) == 0 {
			return false, false, ids.Empty, ErrNoRemainingAuthSignersInWallet
		}
	}
	tx, err := ms.GetWrappedPChainTx()
	if err != nil {
		return false, false, ids.Empty, err
	}
	if err := w.P().Signer().Sign(context.Background(), tx); err != nil {
		return false, false, ids.Empty, fmt.Errorf("error signing tx: %w", err)
	}
	isReady, err := ms.IsReadyToCommit()
	if err != nil {
		return false, false, ids.Empty, err
	}
	if commitIfReady && isReady {
		_, err := w.Commit(ms, waitForTxAcceptanceOnCommit)
		return isReady, err == nil, tx.ID(), err
	}
	return isReady, false, tx.ID(), nil
}

func (w *Wallet) GetRemainingAuthSignersInWallet(ms multisig.Multisig) ([]ids.ShortID, error) {
	_, subnetAuth, err := ms.GetRemainingAuthSigners()
	if err != nil {
		return nil, err
	}
	walletAddrs := w.Addresses()
	subnetAuthInWallet := []ids.ShortID{}
	for _, walletAddr := range walletAddrs {
		for _, addr := range subnetAuth {
			if addr == walletAddr {
				subnetAuthInWallet = append(subnetAuthInWallet, addr)
			}
		}
	}
	return subnetAuthInWallet, nil
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
		return ids.Empty, ErrNotReadyToCommit
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
	for i := 0; i < repeats; i++ {
		ctx, cancel := utils.GetAPILargeContext()
		defer cancel()
		options := []common.Option{common.WithContext(ctx)}
		if !waitForTxAcceptance {
			options = append(options, common.WithAssumeDecided())
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
	// TODO: having a commit error, maybe should be useful to reestart the wallet internal info
	return tx.ID(), issueTxErr
}
