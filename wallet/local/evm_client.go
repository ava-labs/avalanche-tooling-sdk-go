// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"math/big"

	ethereum "github.com/ava-labs/libevm"
	"github.com/ava-labs/libevm/common"
	evmtypes "github.com/ava-labs/libevm/core/types"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
)

// EVMClient represents an EVM client interface for wallet operations
type EVMClient interface {
	// Read Operations (Public)
	GetBalance(ctx context.Context, address common.Address) (*big.Int, error)
	GetChainID(ctx context.Context) (*big.Int, error)
	ContractAlreadyDeployed(ctx context.Context, address string) (bool, error)
	GetContractBytecode(ctx context.Context, address string) ([]byte, error)
	NonceAt(ctx context.Context, address string) (uint64, error)
	BlockNumber(ctx context.Context) (uint64, error)
	CallContract(ctx context.Context, contractAddr common.Address, method string, args ...interface{}) ([]interface{}, error)

	// Transaction Monitoring (Read-only, no private keys needed)
	WaitForTransactionReceipt(ctx context.Context, txHash common.Hash) (*evmtypes.Receipt, error)

	// Gas estimation (read-only)
	EstimateGas(ctx context.Context, callMsg ethereum.CallMsg) (uint64, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	EstimateBaseFee(ctx context.Context) (*big.Int, error)

	Close()
}

// EVM Convenience Methods for LocalWallet
// These methods use the wallet's BuildTx → SignTx → SendTx pattern

// GetEVMBalance gets the balance for an account on EVM
func (w *LocalWallet) GetEVMBalance(ctx context.Context, account account.Account, network network.Network) (*big.Int, error)

// DeployEVMContract deploys a contract using the wallet's transaction pattern
func (w *LocalWallet) DeployEVMContract(ctx context.Context, account account.Account, network network.Network,
	bytecode []byte, abi string, args ...interface{}) (common.Address, common.Hash, uint64, error)

// CallEVMContract calls a contract method using the wallet's transaction pattern
func (w *LocalWallet) CallEVMContract(ctx context.Context, account account.Account, network network.Network,
	contractAddr common.Address, method string, value *big.Int, args ...interface{}) (common.Hash, error)

// ReadEVMContract reads from a contract (no transaction needed)
func (w *LocalWallet) ReadEVMContract(ctx context.Context, network network.Network,
	contractAddr common.Address, method string, args ...interface{}) ([]interface{}, error)

// FundEVMAddress sends ETH from one address to another
func (w *LocalWallet) FundEVMAddress(ctx context.Context, fromAccount account.Account, network network.Network,
	toAddress string, amount *big.Int) (common.Hash, error)

// GetEVMChainID gets the chain ID for EVM
func (w *LocalWallet) GetEVMChainID(ctx context.Context) (*big.Int, error)

// GetEVMBlockNumber gets the current block number
func (w *LocalWallet) GetEVMBlockNumber(ctx context.Context) (uint64, error)

// IsEVMContractDeployed checks if a contract is deployed at the given address
func (w *LocalWallet) IsEVMContractDeployed(ctx context.Context, address string) (bool, error)

// GetEVMContractBytecode gets the bytecode at a contract address
func (w *LocalWallet) GetEVMContractBytecode(ctx context.Context, address string) ([]byte, error)

// GetEVMNonce gets the nonce for an address
func (w *LocalWallet) GetEVMNonce(ctx context.Context, account account.Account, network network.Network) (uint64, error)

// WaitForEVMTransaction waits for a transaction receipt
func (w *LocalWallet) WaitForEVMTransaction(ctx context.Context, txHash common.Hash) (*evmtypes.Receipt, error)
