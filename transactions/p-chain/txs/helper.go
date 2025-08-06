// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	utilsSDK "github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	commonAvago "github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func (d *PublicDeployer) loadCacheWallet(preloadTxs ...ids.ID) (*primary.Wallet, error) {
	var err error
	if d.wallet == nil {
		d.wallet, err = d.loadWallet(preloadTxs...)
	}
	return d.wallet, err
}

func (d *PublicDeployer) loadWallet(subnetIDs ...ids.ID) (*primary.Wallet, error) {
	ctx := context.Background()
	// filter out ids.Empty txs
	filteredTxs := utils.Filter(subnetIDs, func(e ids.ID) bool { return e != ids.Empty })
	wallet, err := primary.MakeWallet(
		ctx,
		d.network.Endpoint,
		d.kc.Keychain,
		secp256k1fx.NewKeychain(),
		primary.WalletConfig{
			SubnetIDs: filteredTxs,
		},
	)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (d *PublicDeployer) Commit(ms multisig.Multisig, wallet wallet.Wallet, waitForTxAcceptance bool) (ids.ID, error) {
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
		issueTxErr = wallet.P().IssueTx(tx, options...)
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
