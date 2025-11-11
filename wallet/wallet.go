// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"math/big"

	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	ethereum "github.com/ava-labs/libevm"
	ethTypes "github.com/ava-labs/libevm/core/types"
)

// Options, Option, and With* functions are defined in options.go

// ContractMethod, Method, and ParseResult are defined in method.go and result.go

// =========================================================================
// Wallet Operations Interfaces
// =========================================================================

// PrimaryOperations handles P/X/C chain operations (Avalanche consensus)
type PrimaryOperations interface {
	// BuildTx constructs a transaction for the specified operation
	BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error)

	// SignTx signs a transaction
	SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error)

	// SendTx submits a signed transaction to the Network
	SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error)

	// SubmitTx is a convenience method that combines BuildTx, SignTx, and SendTx
	SubmitTx(ctx context.Context, params types.SubmitTxParams) (types.SubmitTxResult, error)
}

// Wallet represents the core wallet interface that can be implemented
// by different wallet types (local, API-based, etc.)
type Wallet interface {
	// =========================================================================
	// Network Management
	// =========================================================================

	// SetNetwork sets the default network for wallet operations
	SetNetwork(net network.Network)

	// Network returns the default network for wallet operations
	Network() network.Network

	// =========================================================================
	// Account Management
	// =========================================================================

	// Accounts returns all accounts managed by this wallet with their info
	// Returns map[name]AccountInfo for easy lookup by account name
	Accounts() map[string]account.AccountInfo

	// CreateAccount creates a new account with an optional name.
	// If name is empty, generates a default name (e.g., "account-1")
	// Returns the account info including all chain addresses.
	CreateAccount(name string) (account.AccountInfo, error)

	// ImportAccount imports an account with a name
	// Returns the imported account info
	ImportAccount(name string, spec account.AccountSpec) (account.AccountInfo, error)

	// ExportAccount exports an account by name
	// WARNING: For local accounts, this exposes the private key!
	ExportAccount(name string) (account.AccountSpec, error)

	// Account returns info for a specific account by name
	Account(name string) (account.AccountInfo, error)

	// SetActiveAccount sets the default account for operations
	// Automatically set when first adding an account
	SetActiveAccount(name string) error

	// ActiveAccount returns the currently active account name
	ActiveAccount() string

	// =========================================================================
	// EVM Chain Management
	// =========================================================================

	// SetChain sets the current EVM chain by RPC URL
	// Special case: "c", "cchain", "c-chain" (case-insensitive) uses default network's C-Chain
	// Examples:
	//   - SetChain("C") -> uses network.Endpoint + "/ext/bc/C/rpc"
	//   - SetChain("https://my-l1.example.com/rpc") -> custom L1
	SetChain(rpcURL string) error

	// Chain returns the current EVM chain RPC URL
	// Returns empty string if no chain is set
	Chain() string

	// =========================================================================
	// EVM Operations (Read)
	// =========================================================================

	// EVM operations require SetChain() to be called first
	// All operations return error if no chain is set
	//
	// Operations accept optional Option parameters:
	//   - No options: uses active/default account
	//   - WithAccount("name"): uses named account from wallet
	//   - WithAddress("0x..."): queries arbitrary address

	// Balance returns the balance of an account
	// Examples:
	//   w.Balance()                        // active account
	//   w.Balance(WithAccount("acc1"))     // named account
	//   w.Balance(WithAddress("0x..."))    // any address
	Balance(opts ...Option) (*big.Int, error)

	// ChainID returns the chain ID of the current EVM chain
	ChainID() (*big.Int, error)

	// Nonce returns the nonce for an account
	// Examples:
	//   w.Nonce()                        // active account
	//   w.Nonce(WithAccount("acc1"))     // named account
	//   w.Nonce(WithAddress("0x..."))    // any address
	Nonce(opts ...Option) (uint64, error)

	// BlockNumber returns the latest block number
	BlockNumber() (uint64, error)

	// Block returns block information by number (nil for latest block)
	Block(number *big.Int) (*ethTypes.Block, error)

	// TransactionReceipt retrieves the receipt for a transaction
	TransactionReceipt(txHash common.Hash) (*ethTypes.Receipt, error)

	// Logs returns event logs matching the filter criteria
	Logs(query ethereum.FilterQuery) ([]ethTypes.Log, error)

	// ContractAlreadyDeployed checks if a contract is deployed at the address
	ContractAlreadyDeployed(address string) (bool, error)

	// Code returns the bytecode at the specified address
	Code(address string) ([]byte, error)

	// ReadContract calls a read-only function on a contract
	// Examples:
	//   method := wallet.Method("balanceOf(address)", addr)
	//   w.ReadContract(contractAddr, method)                      // from active account
	//   w.ReadContract(contractAddr, method, WithAccount("acc1")) // from named account
	//   w.ReadContract(contractAddr, method, WithAddress("0x...")) // from any address
	ReadContract(contractAddr common.Address, method ContractMethod, opts ...Option) ([]interface{}, error)

	// =========================================================================
	// EVM Operations (Write)
	// =========================================================================

	// Write operations accept optional Option parameters:
	//   - No options: uses active/default account for signing
	//   - WithAccount("name"): uses named account for signing

	// SignTransaction signs an EVM transaction
	// Examples:
	//   w.SignTransaction(tx)                    // sign with active account
	//   w.SignTransaction(tx, WithAccount("acc1")) // sign with named account
	SignTransaction(tx *ethTypes.Transaction, opts ...Option) (*ethTypes.Transaction, error)

	// SendTransaction sends a signed transaction to the network
	SendTransaction(tx *ethTypes.Transaction) error

	// WaitForTransaction waits for a transaction to be confirmed
	WaitForTransaction(tx *ethTypes.Transaction) (*ethTypes.Receipt, bool, error)

	// FundAddress sends native tokens to the specified address
	// Examples:
	//   w.FundAddress(to, amount)                    // from active account
	//   w.FundAddress(to, amount, WithAccount("acc1")) // from named account
	FundAddress(to string, amount *big.Int, opts ...Option) (*ethTypes.Receipt, error)

	// DeployContract deploys a contract and returns the address, transaction, and receipt
	// Examples:
	//   constructor := wallet.Method("(uint256,address)", protocolVersion, messengerAddr)
	//   w.DeployContract(bin, constructor)                    // deploy from active account
	//   w.DeployContract(bin, constructor, WithAccount("acc1")) // deploy from named account
	DeployContract(binBytes []byte, constructor ContractMethod, opts ...Option) (common.Address, *ethTypes.Transaction, *ethTypes.Receipt, error)

	// WriteContract executes a state-changing transaction to a contract
	// Examples:
	//   method := wallet.Method("setAdmin(address)", adminAddr)
	//   w.WriteContract(contractAddr, nil, method)  // no payment
	//   w.WriteContract(contractAddr, big.NewInt(1000), method)  // with 1000 wei payment
	//   w.WriteContract(contractAddr, nil, method, WithAccount("acc1"))
	//   w.WriteContract(contractAddr, nil, method, WithWarpMessage(warpMsg))  // cross-chain call
	WriteContract(contractAddr common.Address, payment *big.Int, method ContractMethod, opts ...Option) (*ethTypes.Transaction, *ethTypes.Receipt, error)

	// =========================================================================
	// EVM Operations (Utilities)
	// =========================================================================

	// CalculateTxParams calculates gas parameters for a transaction
	// Examples:
	//   w.CalculateTxParams()                     // for active account
	//   w.CalculateTxParams(WithAccount("acc1"))  // for named account
	//   w.CalculateTxParams(WithAddress("0x...")) // for any address
	CalculateTxParams(opts ...Option) (*big.Int, *big.Int, uint64, error)

	// EstimateGasLimit estimates the gas limit for a transaction
	EstimateGasLimit(msg ethereum.CallMsg) (uint64, error)

	// =========================================================================
	// Primary Network Operations (P/X/C Chains)
	// =========================================================================

	// Primary returns the interface for P/X/C chain operations
	// Example: w.Primary().BuildTx(...)
	Primary() PrimaryOperations
}
