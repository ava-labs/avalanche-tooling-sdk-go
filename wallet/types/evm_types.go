// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
)

// EVMDeployContractInput represents EVM contract deployment parameters
type EVMDeployContractInput struct {
	Bytecode []byte
	ABI      string // Method signature like "constructor(string,string,uint256)"
	Args     []interface{}
	Value    *big.Int // ETH to send with deployment
}

func (e *EVMDeployContractInput) GetChainType() string {
	return "evm"
}

func (e *EVMDeployContractInput) Validate() error {
	if len(e.Bytecode) == 0 {
		return fmt.Errorf("bytecode is required")
	}
	if e.ABI == "" {
		return fmt.Errorf("ABI is required")
	}
	return nil
}

// EVMCallContractInput represents EVM contract method calls
type EVMCallContractInput struct {
	ContractAddress common.Address
	Method          string // Method signature like "transfer(address,uint256)"
	Args            []interface{}
	Value           *big.Int // ETH to send with call
}

func (e *EVMCallContractInput) GetChainType() string {
	return "evm"
}

func (e *EVMCallContractInput) Validate() error {
	if e.ContractAddress == (common.Address{}) {
		return fmt.Errorf("contract address is required")
	}
	if e.Method == "" {
		return fmt.Errorf("method is required")
	}
	return nil
}

// EVMTransferInput represents a simple ETH transfer
type EVMTransferInput struct {
	To    string
	Value *big.Int
}

func (e *EVMTransferInput) GetChainType() string {
	return "evm"
}

func (e *EVMTransferInput) Validate() error {
	if e.To == "" {
		return fmt.Errorf("to address is required")
	}
	if e.Value == nil {
		return fmt.Errorf("value is required")
	}
	if e.Value.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("value cannot be negative")
	}
	return nil
}

// EVMReadContractInput represents EVM contract read operations (no transaction)
type EVMReadContractInput struct {
	ContractAddress common.Address
	Method          string
	Args            []interface{}
}

func (e *EVMReadContractInput) GetChainType() string {
	return "evm"
}

func (e *EVMReadContractInput) Validate() error {
	if e.ContractAddress == (common.Address{}) {
		return fmt.Errorf("contract address is required")
	}
	if e.Method == "" {
		return fmt.Errorf("method is required")
	}
	return nil
}

// EVMBuildTxResult represents the result of building an EVM transaction
type EVMBuildTxResult struct {
	Tx *types.Transaction
}

func (e *EVMBuildTxResult) GetChainType() string {
	return "evm"
}

func (e *EVMBuildTxResult) GetTx() interface{} {
	return e.Tx
}

func (e *EVMBuildTxResult) Validate() error {
	if e.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// EVMSendTxResult represents the result of sending an EVM transaction
type EVMSendTxResult struct {
	TxHash          common.Hash
	Receipt         *types.Receipt
	ContractAddress common.Address
	GasUsed         uint64
}

func (e *EVMSendTxResult) GetChainType() string {
	return "evm"
}

func (e *EVMSendTxResult) GetTx() interface{} {
	return e.TxHash
}

func (e *EVMSendTxResult) Validate() error {
	if e.TxHash == (common.Hash{}) {
		return fmt.Errorf("transaction hash cannot be empty")
	}
	return nil
}

func (e *EVMSendTxResult) GetTxHash() common.Hash {
	return e.TxHash
}

func (e *EVMSendTxResult) GetContractAddress() common.Address {
	return e.ContractAddress
}

func (e *EVMSendTxResult) GetGasUsed() uint64 {
	return e.GasUsed
}

// Constructor functions for EVM types
func NewEVMBuildTxResult(tx *types.Transaction) *EVMBuildTxResult {
	return &EVMBuildTxResult{Tx: tx}
}

func NewEVMSendTxResult(txHash common.Hash, receipt *types.Receipt) *EVMSendTxResult {
	result := &EVMSendTxResult{
		TxHash:  txHash,
		Receipt: receipt,
	}

	if receipt != nil {
		result.GasUsed = receipt.GasUsed
		// Extract contract address from receipt if it's a contract deployment
		if len(receipt.ContractAddress) > 0 {
			result.ContractAddress = receipt.ContractAddress
		}
	}

	return result
}
