//go:build query_chain_info
// +build query_chain_info

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
)

const icmMessengerAddr = "0x253b2784c75e510dD0fF1da844684a1aC0aa5fcf"

// This example demonstrates how to query various chain information using the wallet
//
// Environment variables:
//   - PRIVATE_KEY: Private key for the account (optional, only needed for account-specific queries)
//   - EVM_ADDRESS: Address to query balance and nonce (required)
//   - RPC_URL: The RPC URL to connect to (defaults to Fuji C-Chain if not set)
func main() {
	// Get address from environment variable
	evmAddress := os.Getenv("EVM_ADDRESS")
	if evmAddress == "" {
		fmt.Println("Error: EVM_ADDRESS environment variable is required")
		os.Exit(1)
	}

	// Get RPC URL from environment variable, default to Fuji C-Chain
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "C"
	}

	// Create a local wallet with Fuji network
	net := network.FujiNetwork()
	w, err := local.NewLocalWallet(logging.NoLog{}, net)
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	// Import account from private key if provided
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey != "" {
		_, err = w.ImportAccount("user", account.AccountSpec{PrivateKey: privateKey})
		if err != nil {
			fmt.Printf("Failed to import account: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Account imported successfully")
	}

	// Set the EVM chain
	if err := w.SetChain(rpcURL); err != nil {
		fmt.Printf("Failed to set chain: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to: %s\n\n", w.Chain())

	// =========================================================================
	// Chain Information
	// =========================================================================
	fmt.Println("=== Chain Information ===")

	// Get chain ID
	chainID, err := w.ChainID()
	if err != nil {
		fmt.Printf("Failed to get chain ID: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chain ID: %s\n", chainID.String())

	// Get latest block number
	blockNumber, err := w.BlockNumber()
	if err != nil {
		fmt.Printf("Failed to get block number: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Latest Block Number: %d\n", blockNumber)

	// Get latest block details
	block, err := w.Block(nil) // nil = latest block
	if err != nil {
		fmt.Printf("Failed to get block: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Latest Block Hash: %s\n", block.Hash().Hex())
	fmt.Printf("Latest Block Time: %d\n", block.Time())
	fmt.Printf("Latest Block Transactions: %d\n\n", len(block.Transactions()))

	// Get a specific block by number
	if blockNumber > 1 {
		prevBlockNum := new(big.Int).SetUint64(blockNumber - 1)
		specificBlock, err := w.Block(prevBlockNum)
		if err != nil {
			fmt.Printf("Failed to get specific block: %s\n", err)
		} else {
			fmt.Printf("Previous Block Number: %d\n", specificBlock.Number().Uint64())
			fmt.Printf("Previous Block Hash: %s\n\n", specificBlock.Hash().Hex())
		}
	}

	// =========================================================================
	// Account Information (if private key provided)
	// =========================================================================
	fmt.Println("=== Account Information ===")
	if privateKey == "" {
		fmt.Println("No account information available (PRIVATE_KEY environment variable not set)")
		fmt.Println()
	} else {
		accountInfo, err := w.Account("user")
		if err != nil {
			fmt.Printf("Failed to get account info: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Account Address: %s\n", accountInfo.EVMAddress)

		// Get balance
		balance, err := w.Balance()
		if err != nil {
			fmt.Printf("Failed to get balance: %s\n", err)
			os.Exit(1)
		}
		// Convert from wei to AVAX (divide by 10^18)
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		balanceAVAX := new(big.Float).Quo(
			new(big.Float).SetInt(balance),
			new(big.Float).SetInt(divisor),
		)
		fmt.Printf("Balance: %s wei (%s AVAX)\n", balance.String(), balanceAVAX.Text('f', 6))

		// Get nonce
		nonce, err := w.Nonce()
		if err != nil {
			fmt.Printf("Failed to get nonce: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Nonce: %d\n\n", nonce)

		// Get gas parameters
		gasPrice, gasFeeCap, gasLimit, err := w.CalculateTxParams()
		if err != nil {
			fmt.Printf("Failed to calculate tx params: %s\n", err)
		} else {
			fmt.Println("=== Gas Parameters ===")
			fmt.Printf("Gas Price: %s wei\n", gasPrice.String())
			fmt.Printf("Gas Fee Cap: %s wei\n", gasFeeCap.String())
			fmt.Printf("Gas Limit: %d\n\n", gasLimit)
		}
	}

	// =========================================================================
	// Query Arbitrary Address
	// =========================================================================
	fmt.Printf("=== Querying %s ===\n", evmAddress)

	// Get balance
	balance, err := w.Balance(wallet.WithAddress(evmAddress))
	if err != nil {
		fmt.Printf("Failed to get balance: %s\n", err)
	} else {
		// Convert from wei to AVAX (divide by 10^18)
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		balanceAVAX := new(big.Float).Quo(
			new(big.Float).SetInt(balance),
			new(big.Float).SetInt(divisor),
		)
		fmt.Printf("Balance: %s wei (%s AVAX)\n", balance.String(), balanceAVAX.Text('f', 6))
	}

	// Get nonce
	nonce, err := w.Nonce(wallet.WithAddress(evmAddress))
	if err != nil {
		fmt.Printf("Failed to get nonce: %s\n", err)
	} else {
		fmt.Printf("Nonce: %d\n\n", nonce)
	}

	// =========================================================================
	// Contract Code Query
	// =========================================================================
	fmt.Println("=== Contract Code Query ===")

	// Check if an address has contract code (example: ICM Messenger)
	isDeployed, err := w.ContractAlreadyDeployed(icmMessengerAddr)
	if err != nil {
		fmt.Printf("Failed to check if contract deployed: %s\n", err)
	} else {
		fmt.Printf("ICM Messenger (%s) Deployed: %v\n", icmMessengerAddr, isDeployed)
	}

	code, err := w.Code(icmMessengerAddr)
	if err != nil {
		fmt.Printf("Failed to get contract code: %s\n", err)
	} else {
		fmt.Printf("ICM Messenger Code Length: %d bytes\n", len(code))
	}

	fmt.Println("\n✓ Successfully queried chain information")
}
