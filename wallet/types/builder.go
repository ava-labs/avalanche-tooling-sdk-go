// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// BuildTxOutput represents a generic interface for transaction results
type BuildTxOutput interface {
	// GetTxType returns the transaction type identifier
	GetTxType() string
	// GetChainType returns which chain this transaction is for
	GetChainType() string
	// GetTx returns the actual transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// BuildTxParams contains parameters for building transactions
type BuildTxParams struct {
	BaseParams
	BuildTxInput
}

// BuildTxInput represents a generic interface for transaction parameters
type BuildTxInput interface {
	// GetTxType returns the transaction type identifier
	GetTxType() string
	// Validate validates the parameters
	Validate() error
	// GetChainType returns which chain this transaction is for
	GetChainType() string
}

// Validate validates the build transaction parameters
func (p *BuildTxParams) Validate() error {
	if err := p.BaseParams.Validate(); err != nil {
		return err
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

func (p *PChainBuildTxResult) GetTxType() string {
	if p.Tx == nil || p.Tx.Unsigned == nil {
		return "Unknown"
	}
	// Extract tx type from unsigned transaction
	switch p.Tx.Unsigned.(type) {
	case *txs.CreateSubnetTx:
		return "CreateSubnetTx"
	case *txs.ConvertSubnetToL1Tx:
		return "ConvertSubnetToL1Tx"
	case *txs.AddSubnetValidatorTx:
		return "AddSubnetValidatorTx"
	case *txs.RemoveSubnetValidatorTx:
		return "RemoveSubnetValidatorTx"
	case *txs.CreateChainTx:
		return "CreateChainTx"
	case *txs.TransformSubnetTx:
		return "TransformSubnetTx"
	case *txs.AddPermissionlessValidatorTx:
		return "AddPermissionlessValidatorTx"
	case *txs.TransferSubnetOwnershipTx:
		return "TransferSubnetOwnershipTx"
	default:
		return "Unknown"
	}
}

func (p *PChainBuildTxResult) GetChainType() string {
	return "P-Chain"
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

// CChainBuildTxResult represents a C-Chain transaction result
type CChainBuildTxResult struct {
	Tx interface{} // Will be *types.Transaction when C-Chain is implemented
}

func (c *CChainBuildTxResult) GetTxType() string {
	// TODO: Extract tx type from C-Chain transaction when implemented
	return "EVMTransaction"
}

func (c *CChainBuildTxResult) GetChainType() string {
	return "C-Chain"
}

func (c *CChainBuildTxResult) GetTx() interface{} {
	return c.Tx
}

func (c *CChainBuildTxResult) Validate() error {
	if c.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// XChainBuildTxResult represents an X-Chain transaction result
type XChainBuildTxResult struct {
	Tx interface{} // Will be *avm.Tx when X-Chain is implemented
}

func (x *XChainBuildTxResult) GetTxType() string {
	// TODO: Extract tx type from X-Chain transaction when implemented
	return "AVMTransaction"
}

func (x *XChainBuildTxResult) GetChainType() string {
	return "X-Chain"
}

func (x *XChainBuildTxResult) GetTx() interface{} {
	return x.Tx
}

func (x *XChainBuildTxResult) Validate() error {
	if x.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// Constructor functions for each chain type
func NewPChainBuildTxResult(tx *txs.Tx) *PChainBuildTxResult {
	return &PChainBuildTxResult{Tx: tx}
}

func NewCChainBuildTxResult(tx interface{}) *CChainBuildTxResult {
	return &CChainBuildTxResult{Tx: tx}
}

func NewXChainBuildTxResult(tx interface{}) *XChainBuildTxResult {
	return &XChainBuildTxResult{Tx: tx}
}
