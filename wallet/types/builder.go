// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/platformvm/txs"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// BuildTxOutput represents a generic interface for transaction results
type BuildTxOutput interface {
	// GetChainType returns which Avalanche chain this transaction is for.
	// Returns one of: "P-Chain", "X-Chain", "C-Chain"
	// Note: This is different from ChainID (blockchain identifier) or Network (Mainnet/Fuji/etc).
	GetChainType() string
	// GetTx returns the actual transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// BuildTxParams contains parameters for building transactions
type BuildTxParams struct {
	// AccountNames specifies which accounts to use for this transaction.
	// Currently only single-account transactions are supported (first element is used).
	// Future: Will support multi-account for multisig transactions.
	AccountNames []string
	BuildTxInput
}

// BuildTxInput represents a generic interface for transaction parameters
type BuildTxInput interface {
	// Validate validates the parameters
	Validate() error
	// GetChainType returns which Avalanche chain this transaction is for.
	// Returns one of: "P-Chain", "X-Chain", "C-Chain"
	// Note: This is different from ChainID (blockchain identifier) or Network (Mainnet/Fuji/etc).
	GetChainType() string
}

// Validate validates the build transaction parameters
func (p *BuildTxParams) Validate() error {
	// TODO: Support multiple accounts for multisig transactions
	if len(p.AccountNames) > 1 {
		return fmt.Errorf("only one account name is currently supported")
	}
	if p.BuildTxInput == nil {
		return fmt.Errorf("build tx input is required")
	}
	return p.BuildTxInput.Validate()
}

// BuildTxResult represents the result of building a transaction
type BuildTxResult struct {
	BuildTxOutput
}

// Validate validates the build transaction result
func (r *BuildTxResult) Validate() error {
	if r.BuildTxOutput == nil {
		return fmt.Errorf("build tx output is required")
	}
	return r.BuildTxOutput.Validate()
}

// PChainBuildTxResult represents a P-Chain transaction result
type PChainBuildTxResult struct {
	Tx *txs.Tx
}

func (p *PChainBuildTxResult) GetChainType() string {
	return constants.ChainTypePChain
}

func (p *PChainBuildTxResult) GetTx() interface{} {
	return p.Tx
}

func (p *PChainBuildTxResult) Validate() error {
	if p.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// Constructor functions for each chain type
func NewPChainBuildTxResult(tx *txs.Tx) *PChainBuildTxResult {
	return &PChainBuildTxResult{Tx: tx}
}
