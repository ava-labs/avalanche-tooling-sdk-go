// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// BuildTxOutput represents a generic interface for transaction results
type BuildTxOutput interface {
	// GetChainType returns which chain this transaction is for
	GetChainType() string
	// GetTx returns the actual transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// BuildTxParams contains parameters for building transactions
type BuildTxParams struct {
	AccountNames []string
	BuildTxInput
}

// BuildTxInput represents a generic interface for transaction parameters
type BuildTxInput interface {
	// Validate validates the parameters
	Validate() error
	// GetChainType returns which chain this transaction is for
	GetChainType() string
}

// Validate validates the build transaction parameters
func (p *BuildTxParams) Validate() error {
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
	return ChainTypePChain
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
