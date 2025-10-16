// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/platformvm/txs"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
)

// SendTxOutput represents a generic interface for sent transaction results
type SendTxOutput interface {
	// GetChainType returns which chain this transaction is for
	GetChainType() string
	// GetTx returns the actual sent transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// SendTxParams contains parameters for sending transactions
type SendTxParams struct {
	Account account.Account
	Network network.Network
	*SignTxResult
}

// Validate validates the send transaction parameters
func (p *SendTxParams) Validate() error {
	if p.Account == nil {
		return fmt.Errorf("account is required")
	}
	if p.Network.Kind == network.Undefined {
		return fmt.Errorf("network is required")
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
