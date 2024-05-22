// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"time"
)

func Commit(
	wallet primary.Wallet,
	tx *txs.Tx,
	waitForTxAcceptance bool,
) (ids.ID, error) {
	const (
		repeats             = 3
		sleepBetweenRepeats = 2 * time.Second
	)
	var issueTxErr error
	//wallet, err := d.loadCacheWallet()
	//if err != nil {
	//	return ids.Empty, err
	//}
	for i := 0; i < repeats; i++ {
		ctx, cancel := utils.GetAPILargeContext()
		defer cancel()
		options := []common.Option{common.WithContext(ctx)}
		if !waitForTxAcceptance {
			options = append(options, common.WithAssumeDecided())
		}
		issueTxErr = wallet.P().IssueTx(tx, options...)
		if issueTxErr == nil {
			break
		}
		if ctx.Err() != nil {
			issueTxErr = fmt.Errorf("timeout issuing/verifying tx with ID %s: %w", tx.ID(), issueTxErr)
		} else {
			issueTxErr = fmt.Errorf("error issuing tx with ID %s: %w", tx.ID(), issueTxErr)
		}
		ux.Logger.RedXToUser("%s", issueTxErr)
		time.Sleep(sleepBetweenRepeats)
	}
	if issueTxErr != nil {
		d.cleanCacheWallet()
	}
	return tx.ID(), issueTxErr
}
