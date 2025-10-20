// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"math/big"

	ethereum "github.com/ava-labs/libevm"
	"github.com/ava-labs/libevm/common"
	evmtypes "github.com/ava-labs/libevm/core/types"
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
