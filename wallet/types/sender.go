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
	// GetTxType returns the transaction type identifier
	GetTxType() string
	// GetChainType returns which chain this transaction is for
	GetChainType() string
	// GetTx returns the actual sent transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// SendTxParams contains parameters for sending transactions
type SendTxParams struct {
	BaseParams
	*SignTxResult
}

// Validate validates the send transaction parameters
func (p *SendTxParams) Validate() error {
	if err := p.BaseParams.Validate(); err != nil {
		return err
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

func (p *PChainSendTxResult) GetTxType() string {
	if p.Tx == nil || p.Tx.Unsigned == nil {
		return constants.TxTypeUnknown
	}
	// Extract tx type from unsigned transaction
	switch p.Tx.Unsigned.(type) {
	case *txs.CreateSubnetTx:
		return constants.PChainCreateSubnetTx
	case *txs.ConvertSubnetToL1Tx:
		return constants.PChainConvertSubnetToL1Tx
	case *txs.AddSubnetValidatorTx:
		return constants.PChainAddSubnetValidatorTx
	case *txs.RemoveSubnetValidatorTx:
		return constants.PChainRemoveSubnetValidatorTx
	case *txs.CreateChainTx:
		return constants.PChainCreateChainTx
	case *txs.TransformSubnetTx:
		return constants.PChainTransformSubnetTx
	case *txs.AddPermissionlessValidatorTx:
		return constants.PChainAddPermissionlessValidatorTx
	case *txs.TransferSubnetOwnershipTx:
		return constants.PChainTransferSubnetOwnershipTx
	default:
		return constants.TxTypeUnknown
	}
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

// CChainSendTxResult represents a C-Chain sent transaction result
type CChainSendTxResult struct {
	Tx interface{} // Will be *types.Transaction when C-Chain is implemented
}

func (c *CChainSendTxResult) GetTxType() string {
	// TODO: Extract tx type from C-Chain transaction when implemented
	return constants.TxTypeEVMTransaction
}

func (c *CChainSendTxResult) GetChainType() string {
	return constants.ChainTypeCChain
}

func (c *CChainSendTxResult) GetTx() interface{} {
	return c.Tx
}

func (c *CChainSendTxResult) Validate() error {
	if c.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// XChainSendTxResult represents an X-Chain sent transaction result
type XChainSendTxResult struct {
	Tx interface{} // Will be *avm.Tx when X-Chain is implemented
}

func (x *XChainSendTxResult) GetTxType() string {
	// TODO: Extract tx type from X-Chain transaction when implemented
	return constants.TxTypeAVMTransaction
}

func (x *XChainSendTxResult) GetChainType() string {
	return constants.ChainTypeXChain
}

func (x *XChainSendTxResult) GetTx() interface{} {
	return x.Tx
}

func (x *XChainSendTxResult) Validate() error {
	if x.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// Constructor functions for SendTxResult
func NewPChainSendTxResult(tx *txs.Tx) *PChainSendTxResult {
	return &PChainSendTxResult{Tx: tx}
}

func NewCChainSendTxResult(tx interface{}) *CChainSendTxResult {
	return &CChainSendTxResult{Tx: tx}
}

func NewXChainSendTxResult(tx interface{}) *XChainSendTxResult {
	return &XChainSendTxResult{Tx: tx}
}
