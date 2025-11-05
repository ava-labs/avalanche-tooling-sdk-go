// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package pchain

import (
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// sendWithRetry submits a transaction to the network with retry logic
func sendWithRetry(wallet *primary.Wallet, tx *txs.Tx) error {
	const (
		repeats             = 3
		sleepBetweenRepeats = 2 * time.Second
	)
	var issueTxErr error
	for i := 0; i < repeats; i++ {
		ctx, cancel := utils.GetAPILargeContext()
		defer cancel()
		options := []common.Option{common.WithContext(ctx)}
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
		return fmt.Errorf("issue tx error %w", issueTxErr)
	}
	return nil
}

// SendTx sends P-Chain transactions
func SendTx(wallet *primary.Wallet, params types.SendTxParams) (types.SendTxResult, error) {
	// Validate transaction is defined
	if params.SignTxResult.Undefined() {
		return types.SendTxResult{}, multisig.ErrUndefinedTx
	}

	// Validate transaction is ready to commit
	isReady, err := params.SignTxResult.IsReadyToCommit()
	if err != nil {
		return types.SendTxResult{}, err
	}
	if !isReady {
		return types.SendTxResult{}, errors.New("tx is not fully signed so can't be committed")
	}

	// Extract the P-Chain transaction
	tx, err := params.SignTxResult.GetWrappedPChainTx()
	if err != nil {
		return types.SendTxResult{}, err
	}

	// Submit the signed transaction to the network with retry logic
	if err := sendWithRetry(wallet, tx); err != nil {
		return types.SendTxResult{}, fmt.Errorf("error sending tx: %w", err)
	}

	pChainSendResult := types.NewPChainSendTxResult(tx)
	return types.SendTxResult{SendTxOutput: pChainSendResult}, nil
}
