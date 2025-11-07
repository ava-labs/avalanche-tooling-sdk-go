// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"errors"
	"math/big"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	ethereum "github.com/ava-labs/libevm"
	ethCommon "github.com/ava-labs/libevm/common"
	ethTypes "github.com/ava-labs/libevm/core/types"
)

// SetChain sets the current EVM chain by RPC URL
func (w *LocalWallet) SetChain(rpcURL string) error {
	_ = rpcURL
	return errors.New("not implemented")
}

// Chain returns the current EVM chain RPC URL
func (w *LocalWallet) Chain() string {
	return ""
}

// Balance returns the balance of an account
func (w *LocalWallet) Balance(opts ...wallet.Option) (*big.Int, error) {
	_ = opts
	return nil, errors.New("not implemented")
}

// ChainID returns the chain ID of the current EVM chain
func (w *LocalWallet) ChainID() (*big.Int, error) {
	return nil, errors.New("not implemented")
}

// Nonce returns the nonce for an account
func (w *LocalWallet) Nonce(opts ...wallet.Option) (uint64, error) {
	_ = opts
	return 0, errors.New("not implemented")
}

// BlockNumber returns the latest block number
func (w *LocalWallet) BlockNumber() (uint64, error) {
	return 0, errors.New("not implemented")
}

// Block returns block information by number
func (w *LocalWallet) Block(number *big.Int) (*ethTypes.Block, error) {
	_ = number
	return nil, errors.New("not implemented")
}

// TransactionReceipt retrieves the receipt for a transaction
func (w *LocalWallet) TransactionReceipt(txHash ethCommon.Hash) (*ethTypes.Receipt, error) {
	_ = txHash
	return nil, errors.New("not implemented")
}

// Logs returns event logs matching the filter criteria
func (w *LocalWallet) Logs(query ethereum.FilterQuery) ([]ethTypes.Log, error) {
	_ = query
	return nil, errors.New("not implemented")
}

// ContractAlreadyDeployed checks if a contract is deployed at the address
func (w *LocalWallet) ContractAlreadyDeployed(address string) (bool, error) {
	_ = address
	return false, errors.New("not implemented")
}

// Code returns the bytecode at the specified address
func (w *LocalWallet) Code(address string) ([]byte, error) {
	_ = address
	return nil, errors.New("not implemented")
}

// ReadContract calls a read-only function on a contract
func (w *LocalWallet) ReadContract(contractAddr ethCommon.Address, method wallet.ContractMethod, opts ...wallet.Option) ([]interface{}, error) {
	_, _, _ = contractAddr, method, opts
	return nil, errors.New("not implemented")
}

// SignTransaction signs an EVM transaction
func (w *LocalWallet) SignTransaction(tx *ethTypes.Transaction, opts ...wallet.Option) (*ethTypes.Transaction, error) {
	_, _ = tx, opts
	return nil, errors.New("not implemented")
}

// SendTransaction sends a signed transaction to the network
func (w *LocalWallet) SendTransaction(tx *ethTypes.Transaction) error {
	_ = tx
	return errors.New("not implemented")
}

// WaitForTransaction waits for a transaction to be confirmed
func (w *LocalWallet) WaitForTransaction(tx *ethTypes.Transaction) (*ethTypes.Receipt, bool, error) {
	_ = tx
	return nil, false, errors.New("not implemented")
}

// FundAddress sends native tokens to the specified address
func (w *LocalWallet) FundAddress(to string, amount *big.Int, opts ...wallet.Option) (*ethTypes.Receipt, error) {
	_, _, _ = to, amount, opts
	return nil, errors.New("not implemented")
}

// DeployContract deploys a contract and returns the address, transaction, and receipt
func (w *LocalWallet) DeployContract(binBytes []byte, constructor wallet.ContractMethod, opts ...wallet.Option) (ethCommon.Address, *ethTypes.Transaction, *ethTypes.Receipt, error) {
	_, _, _ = binBytes, constructor, opts
	return ethCommon.Address{}, nil, nil, errors.New("not implemented")
}

// WriteContract executes a state-changing transaction to a contract
func (w *LocalWallet) WriteContract(contractAddr ethCommon.Address, payment *big.Int, method wallet.ContractMethod, opts ...wallet.Option) (*ethTypes.Transaction, *ethTypes.Receipt, error) {
	_, _, _, _ = contractAddr, payment, method, opts
	return nil, nil, errors.New("not implemented")
}

// CalculateTxParams calculates gas parameters for a transaction
func (w *LocalWallet) CalculateTxParams(opts ...wallet.Option) (*big.Int, *big.Int, uint64, error) {
	_ = opts
	return nil, nil, 0, errors.New("not implemented")
}

// EstimateGasLimit estimates the gas limit for a transaction
func (w *LocalWallet) EstimateGasLimit(msg ethereum.CallMsg) (uint64, error) {
	_ = msg
	return 0, errors.New("not implemented")
}
