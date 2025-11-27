// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/platformvm/txs"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// SendTxOutput represents a generic interface for sent transaction results
type SendTxOutput interface {
	// GetChainType returns which Avalanche chain this transaction is for.
	// Returns one of: "P-Chain", "X-Chain", "C-Chain"
	// Note: This is different from ChainID (blockchain identifier) or Network (Mainnet/Fuji/etc).
	GetChainType() string
	// GetTx returns the actual sent transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// SendTxParams contains parameters for sending transactions
type SendTxParams struct {
	// AccountNames specifies which accounts to use for sending this transaction.
	// Currently only single-account transactions are supported (first element is used).
	// Future: Will support multi-account for multisig transactions.
	AccountNames []string
	*SignTxResult
}

// Validate validates the send transaction parameters
func (p *SendTxParams) Validate() error {
	// TODO: Support multiple accounts for multisig transactions
	if len(p.AccountNames) > 1 {
		return fmt.Errorf("only one account name is currently supported")
	}
	if p.SignTxResult == nil {
		return fmt.Errorf("sign tx result is required")
	}
	return p.SignTxResult.Validate()
}

// SendTxResult represents the result of sending a transaction
type SendTxResult struct {
	SendTxOutput
}

// Validate validates the send transaction result
func (r *SendTxResult) Validate() error {
	if r.SendTxOutput == nil {
		return fmt.Errorf("send tx output is required")
	}
	return r.SendTxOutput.Validate()
}

// PChainSendTxResult represents a P-Chain sent transaction result
type PChainSendTxResult struct {
	Tx *txs.Tx
}

func (p *PChainSendTxResult) GetChainType() string {
	return constants.ChainTypePChain
}

func (p *PChainSendTxResult) GetTx() interface{} {
	return p.Tx
}

func (p *PChainSendTxResult) Validate() error {
	if p.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// Constructor functions for SendTxResult
func NewPChainSendTxResult(tx *txs.Tx) *PChainSendTxResult {
	return &PChainSendTxResult{Tx: tx}
}
