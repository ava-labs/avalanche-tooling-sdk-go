// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"

	ethereum "github.com/ava-labs/libevm"
)

// SimulateTransaction simulates a transaction that has already been mined to extract error information.
// Uses standard eth_call at the block before the transaction was mined, avoiding debug methods.
// Returns the error from the simulation which contains revert data if the transaction reverted.
func SimulateTransaction(
	rpcURL string,
	txHash string,
) error {
	client, err := GetClient(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	defer client.Close()

	// Get the transaction
	tx, isPending, err := client.TransactionByHash(common.HexToHash(txHash))
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if isPending {
		return fmt.Errorf("transaction is still pending")
	}

	// Get the receipt to find the block number
	receipt, err := client.TransactionReceipt(common.HexToHash(txHash))
	if err != nil {
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	// If transaction succeeded, no error to simulate
	if receipt.Status == 1 {
		return nil
	}

	// Recover the sender address from the transaction
	chainID, err := client.GetChainID()
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	signer := types.NewLondonSigner(chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("failed to recover sender address: %w", err)
	}

	// Construct the call message from the transaction
	msg := ethereum.CallMsg{
		From:      from,
		To:        tx.To(),
		Gas:       tx.Gas(),
		GasPrice:  tx.GasPrice(),
		GasFeeCap: tx.GasFeeCap(),
		GasTipCap: tx.GasTipCap(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	}

	// Simulate at the block before the transaction was mined
	blockNumber := new(big.Int).SetUint64(receipt.BlockNumber.Uint64() - 1)
	_, callErr := client.CallContract(msg, blockNumber)

	if callErr != nil {
		return callErr
	}

	// If simulation succeeded but original tx failed, it's a race condition or state issue
	return fmt.Errorf("transaction failed on-chain but simulation succeeded (possible race condition)")
}
