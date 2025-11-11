// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/ava-labs/avalanchego/vms/platformvm/warp"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm/contract"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	ethereum "github.com/ava-labs/libevm"
	ethCommon "github.com/ava-labs/libevm/common"
	ethTypes "github.com/ava-labs/libevm/core/types"
)

// SetChain sets the current EVM chain by RPC URL
// Special case: "c", "cchain", "c-chain" (case-insensitive) uses default network's C-Chain
func (w *LocalWallet) SetChain(rpcURL string) error {
	// Close existing client if any
	if w.evmClient != nil {
		w.evmClient.Close()
		w.evmClient = nil
		w.evmRPC = ""
	}

	// Validate non-empty
	rpcURL = strings.TrimSpace(rpcURL)
	if rpcURL == "" {
		return fmt.Errorf("RPC URL cannot be empty")
	}

	// Handle C-Chain shortcuts (case-insensitive)
	lowerRPC := strings.ToLower(rpcURL)
	if lowerRPC == "c" || lowerRPC == "cchain" || lowerRPC == "c-chain" {
		rpcURL = w.defaultNetwork.Endpoint + "/ext/bc/C/rpc"
	} else {
		// Validate URL format for non-C-Chain cases
		parsedURL, err := url.Parse(rpcURL)
		if err != nil {
			return fmt.Errorf("invalid RPC URL format %q: %w", rpcURL, err)
		}
		// Ensure it looks like a valid URL (has scheme or host)
		if parsedURL.Scheme == "" && parsedURL.Host == "" {
			return fmt.Errorf("invalid RPC URL %q: must be a valid URL", rpcURL)
		}
	}

	// Connect to the EVM RPC endpoint
	client, err := evm.GetClient(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to EVM chain at %s: %w", rpcURL, err)
	}

	w.evmClient = &client
	w.evmRPC = rpcURL
	return nil
}

// Chain returns the current EVM chain RPC URL
func (w *LocalWallet) Chain() string {
	return w.evmRPC
}

// ensureChainSet returns an error if no EVM chain is currently set
func (w *LocalWallet) ensureChainSet() error {
	if w.evmClient == nil || w.evmRPC == "" {
		return fmt.Errorf("EVM chain not set, call SetChain() first")
	}
	return nil
}

// getAccount returns the account to use based on options or active account
func (w *LocalWallet) getAccount(opts ...wallet.Option) (account.Account, error) {
	options := &wallet.Options{}
	for _, opt := range opts {
		opt(options)
	}

	var accountName string
	if options.AccountName != "" {
		accountName = options.AccountName
	} else {
		if w.activeAccount == "" {
			return nil, fmt.Errorf("no active account set")
		}
		accountName = w.activeAccount
	}

	acc, exists := w.accounts[accountName]
	if !exists {
		return nil, fmt.Errorf("account %q not found", accountName)
	}
	return acc, nil
}

// resolveAddress resolves an address from options or active account
func (w *LocalWallet) resolveAddress(opts ...wallet.Option) (string, error) {
	options := &wallet.Options{}
	for _, opt := range opts {
		opt(options)
	}

	// If explicit address provided, use it
	if options.Address != "" {
		return options.Address, nil
	}

	// Get account and return its EVM address
	acc, err := w.getAccount(opts...)
	if err != nil {
		return "", err
	}
	return acc.GetEVMAddress()
}

// getSigner returns an EVM signer for the specified account or active account
func (w *LocalWallet) getSigner(opts ...wallet.Option) (*evm.Signer, error) {
	acc, err := w.getAccount(opts...)
	if err != nil {
		return nil, err
	}

	// Get keychain from account
	keychain, err := acc.GetKeychain()
	if err != nil {
		return nil, fmt.Errorf("failed to get keychain: %w", err)
	}

	// Create EVM signer
	signer, err := evm.NewSigner(keychain)
	if err != nil {
		return nil, fmt.Errorf("failed to create EVM signer: %w", err)
	}

	return signer, nil
}

// Balance returns the balance of an account
func (w *LocalWallet) Balance(opts ...wallet.Option) (*big.Int, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	address, err := w.resolveAddress(opts...)
	if err != nil {
		return nil, err
	}

	return w.evmClient.GetAddressBalance(address)
}

// ChainID returns the chain ID of the current EVM chain
func (w *LocalWallet) ChainID() (*big.Int, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	return w.evmClient.GetChainID()
}

// Nonce returns the nonce for an account
func (w *LocalWallet) Nonce(opts ...wallet.Option) (uint64, error) {
	if err := w.ensureChainSet(); err != nil {
		return 0, err
	}

	address, err := w.resolveAddress(opts...)
	if err != nil {
		return 0, err
	}

	return w.evmClient.NonceAt(address)
}

// BlockNumber returns the latest block number
func (w *LocalWallet) BlockNumber() (uint64, error) {
	if err := w.ensureChainSet(); err != nil {
		return 0, err
	}

	return w.evmClient.BlockNumber()
}

// Block returns block information by number
func (w *LocalWallet) Block(number *big.Int) (*ethTypes.Block, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	return w.evmClient.BlockByNumber(number)
}

// TransactionReceipt retrieves the receipt for a transaction
func (w *LocalWallet) TransactionReceipt(txHash ethCommon.Hash) (*ethTypes.Receipt, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	return w.evmClient.TransactionReceipt(txHash)
}

// Logs returns event logs matching the filter criteria
func (w *LocalWallet) Logs(query ethereum.FilterQuery) ([]ethTypes.Log, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	return w.evmClient.FilterLogs(query)
}

// ContractAlreadyDeployed checks if a contract is deployed at the address
func (w *LocalWallet) ContractAlreadyDeployed(address string) (bool, error) {
	if err := w.ensureChainSet(); err != nil {
		return false, err
	}

	return w.evmClient.ContractAlreadyDeployed(address)
}

// Code returns the bytecode at the specified address
func (w *LocalWallet) Code(address string) ([]byte, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	return w.evmClient.GetContractBytecode(address)
}

// ReadContract calls a read-only function on a contract
func (w *LocalWallet) ReadContract(contractAddr ethCommon.Address, method wallet.ContractMethod, opts ...wallet.Option) ([]interface{}, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	// Resolve the from address from options
	from := ethCommon.Address{}
	if fromAddr, err := w.resolveAddress(opts...); err == nil {
		from = ethCommon.HexToAddress(fromAddr)
	}

	return contract.CallToMethod(w.evmRPC, from, contractAddr, method.Spec, nil, method.Params...)
}

// SignTransaction signs an EVM transaction
func (w *LocalWallet) SignTransaction(tx *ethTypes.Transaction, opts ...wallet.Option) (*ethTypes.Transaction, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	signer, err := w.getSigner(opts...)
	if err != nil {
		return nil, err
	}

	chainID, err := w.evmClient.GetChainID()
	if err != nil {
		return nil, err
	}

	return signer.SignTx(chainID, tx)
}

// SendTransaction sends a signed transaction to the network
func (w *LocalWallet) SendTransaction(tx *ethTypes.Transaction) error {
	if err := w.ensureChainSet(); err != nil {
		return err
	}

	return w.evmClient.SendTransaction(tx)
}

// WaitForTransaction waits for a transaction to be confirmed
func (w *LocalWallet) WaitForTransaction(tx *ethTypes.Transaction) (*ethTypes.Receipt, bool, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, false, err
	}

	return w.evmClient.WaitForTransaction(tx)
}

// FundAddress sends native tokens to the specified address
func (w *LocalWallet) FundAddress(to string, amount *big.Int, opts ...wallet.Option) (*ethTypes.Receipt, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, err
	}

	signer, err := w.getSigner(opts...)
	if err != nil {
		return nil, err
	}

	return w.evmClient.FundAddress(signer, to, amount)
}

// DeployContract deploys a contract and returns the address, transaction, and receipt
func (w *LocalWallet) DeployContract(binBytes []byte, constructor wallet.ContractMethod, opts ...wallet.Option) (ethCommon.Address, *ethTypes.Transaction, *ethTypes.Receipt, error) {
	if err := w.ensureChainSet(); err != nil {
		return ethCommon.Address{}, nil, nil, err
	}

	signer, err := w.getSigner(opts...)
	if err != nil {
		return ethCommon.Address{}, nil, nil, err
	}

	return contract.DeployContract(w.evmRPC, signer, binBytes, constructor.Spec, constructor.Params...)
}

// WriteContract executes a state-changing transaction to a contract
func (w *LocalWallet) WriteContract(contractAddr ethCommon.Address, payment *big.Int, method wallet.ContractMethod, opts ...wallet.Option) (*ethTypes.Transaction, *ethTypes.Receipt, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, nil, err
	}

	signer, err := w.getSigner(opts...)
	if err != nil {
		return nil, nil, err
	}

	// Process options
	options := &wallet.Options{}
	for _, opt := range opts {
		opt(options)
	}

	// Handle warp message if provided
	if options.WarpMessage != nil {
		warpMsg, ok := options.WarpMessage.(*warp.Message)
		if !ok {
			return nil, nil, fmt.Errorf("invalid warp message type")
		}
		return contract.TxToMethodWithWarpMessage(
			w.logger,
			w.evmRPC,
			signer,
			contractAddr,
			warpMsg,
			payment,
			method.Name(),
			options.ErrorMap,
			method.Spec,
			method.Params...,
		)
	}

	// Regular contract call
	return contract.TxToMethod(
		w.logger,
		w.evmRPC,
		signer,
		contractAddr,
		payment,
		method.Name(),
		options.ErrorMap,
		method.Spec,
		method.Params...,
	)
}

// CalculateTxParams calculates gas parameters for a transaction
func (w *LocalWallet) CalculateTxParams(opts ...wallet.Option) (*big.Int, *big.Int, uint64, error) {
	if err := w.ensureChainSet(); err != nil {
		return nil, nil, 0, err
	}

	address, err := w.resolveAddress(opts...)
	if err != nil {
		return nil, nil, 0, err
	}

	return w.evmClient.CalculateTxParams(address)
}

// EstimateGasLimit estimates the gas limit for a transaction
func (w *LocalWallet) EstimateGasLimit(msg ethereum.CallMsg) (uint64, error) {
	if err := w.ensureChainSet(); err != nil {
		return 0, err
	}

	return w.evmClient.EstimateGasLimit(msg)
}
