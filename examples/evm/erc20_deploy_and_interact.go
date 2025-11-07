//go:build erc20_deploy_and_interact
// +build erc20_deploy_and_interact

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"

	_ "embed"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
)

//go:embed erc20.bin
var erc20Bin string

// This example demonstrates how to deploy an ERC20 token contract and interact with it
//
// Environment variables:
//   - PRIVATE_KEY: Private key for deploying and interacting with the contract (required)
//   - RPC_URL: The RPC URL to connect to (defaults to Fuji C-Chain if not set)

// Helper function to read and parse ERC20 balance
func getBalance(w wallet.Wallet, tokenAddr common.Address, address string) (*big.Int, error) {
	method := wallet.Method("balanceOf(address)->(uint256)", common.HexToAddress(address))
	result, err := w.ReadContract(tokenAddr, method)
	if err != nil {
		return nil, err
	}
	return wallet.ParseResult[*big.Int](result)
}

func main() {
	// Get private key from environment
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		fmt.Println("Error: PRIVATE_KEY environment variable is required")
		os.Exit(1)
	}

	// Get RPC URL from environment variable, default to Fuji C-Chain
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "C"
	}

	// Create wallet with Fuji network
	net := network.FujiNetwork()
	w, err := local.NewLocalWallet(logging.NoLog{}, net)
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	// Import account from private key
	_, err = w.ImportAccount("deployer", account.AccountSpec{PrivateKey: privateKey})
	if err != nil {
		fmt.Printf("Failed to import account: %s\n", err)
		os.Exit(1)
	}

	// Set the EVM chain
	if err := w.SetChain(rpcURL); err != nil {
		fmt.Printf("Failed to set chain: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deploying and interacting with ERC20 token on: %s\n\n", w.Chain())

	// Get deployer address
	accountInfo, err := w.Account("deployer")
	if err != nil {
		fmt.Printf("Failed to get account info: %s\n", err)
		os.Exit(1)
	}
	deployerAddr := accountInfo.EVMAddress
	fmt.Printf("Deployer address: %s\n", deployerAddr)

	// Check balance before deployment
	balance, err := w.Balance()
	if err != nil {
		fmt.Printf("Failed to get balance: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Deployer balance: %s wei\n\n", balance.String())

	// Deploy ERC20 token contract
	// Constructor: constructor(string memory symbol, address funded, uint256 balance)
	fmt.Println("=== Deploying ERC20 Token ===")
	tokenSymbol := "TEST"
	initialSupply := big.NewInt(1000000) // 1 million tokens (before decimals)

	constructor := wallet.Method(
		"(string,address,uint256)",
		tokenSymbol,
		common.HexToAddress(deployerAddr),
		initialSupply,
	)

	tokenAddr, tx, receipt, err := w.DeployContract([]byte(erc20Bin), constructor)
	if err != nil {
		fmt.Printf("Failed to deploy contract: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Token deployed at: %s\n", tokenAddr.Hex())
	fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
	fmt.Printf("Gas used: %d\n\n", receipt.GasUsed)

	// Read token information
	fmt.Println("=== Reading Token Information ===")

	// Get token name
	method := wallet.Method("name()->(string)")
	result, err := w.ReadContract(tokenAddr, method)
	if err != nil {
		fmt.Printf("Failed to read name: %s\n", err)
		os.Exit(1)
	}
	name, err := wallet.ParseResult[string](result)
	if err != nil {
		fmt.Printf("Failed to parse name: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Name: %s\n", name)

	// Get token symbol
	method = wallet.Method("symbol()->(string)")
	result, err = w.ReadContract(tokenAddr, method)
	if err != nil {
		fmt.Printf("Failed to read symbol: %s\n", err)
		os.Exit(1)
	}
	symbol, err := wallet.ParseResult[string](result)
	if err != nil {
		fmt.Printf("Failed to parse symbol: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Symbol: %s\n", symbol)

	// Get decimals
	method = wallet.Method("decimals()->(uint8)")
	result, err = w.ReadContract(tokenAddr, method)
	if err != nil {
		fmt.Printf("Failed to read decimals: %s\n", err)
		os.Exit(1)
	}
	decimals, err := wallet.ParseResult[uint8](result)
	if err != nil {
		fmt.Printf("Failed to parse decimals: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Decimals: %d\n", decimals)

	// Get total supply
	method = wallet.Method("totalSupply()->(uint256)")
	result, err = w.ReadContract(tokenAddr, method)
	if err != nil {
		fmt.Printf("Failed to read totalSupply: %s\n", err)
		os.Exit(1)
	}
	totalSupply, err := wallet.ParseResult[*big.Int](result)
	if err != nil {
		fmt.Printf("Failed to parse totalSupply: %s\n", err)
		os.Exit(1)
	}
	// Convert to human-readable format (divide by 10^decimals)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	totalSupply = new(big.Int).Div(totalSupply, divisor)
	fmt.Printf("Total Supply: %s %s\n", totalSupply.String(), symbol)

	// Get balance of deployer
	deployerBalance, err := getBalance(w, tokenAddr, deployerAddr)
	if err != nil {
		fmt.Printf("Failed to read deployer balance: %s\n", err)
		os.Exit(1)
	}
	deployerBalance = new(big.Int).Div(deployerBalance, divisor)
	fmt.Printf("Deployer Balance: %s %s\n\n", deployerBalance.String(), symbol)

	// Transfer tokens to another address
	fmt.Println("=== Transferring Tokens ===")
	recipientAddr := "0x0000000000000000000000000000000000000001" // Example recipient
	transferAmount := big.NewInt(100)                             // Transfer 100 tokens
	transferAmountRaw := new(big.Int).Mul(transferAmount, divisor)
	fmt.Printf("Transferring %s %s to %s\n", transferAmount.String(), symbol, recipientAddr)

	method = wallet.Method("transfer(address,uint256)->(bool)", common.HexToAddress(recipientAddr), transferAmountRaw)
	tx, receipt, err = w.WriteContract(tokenAddr, nil, method)
	if err != nil {
		fmt.Printf("Failed to transfer tokens: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
	fmt.Printf("Gas used: %d\n", receipt.GasUsed)
	fmt.Printf("Status: %d (1 = success)\n\n", receipt.Status)

	// Verify balances after transfer
	fmt.Println("=== Verifying Balances After Transfer ===")

	// Check deployer balance
	deployerBalance, err = getBalance(w, tokenAddr, deployerAddr)
	if err != nil {
		fmt.Printf("Failed to read deployer balance: %s\n", err)
		os.Exit(1)
	}
	deployerBalance = new(big.Int).Div(deployerBalance, divisor)
	fmt.Printf("Deployer Balance: %s %s\n", deployerBalance.String(), symbol)

	// Check recipient balance
	recipientBalance, err := getBalance(w, tokenAddr, recipientAddr)
	if err != nil {
		fmt.Printf("Failed to read recipient balance: %s\n", err)
		os.Exit(1)
	}
	recipientBalance = new(big.Int).Div(recipientBalance, divisor)
	fmt.Printf("Recipient Balance: %s %s\n\n", recipientBalance.String(), symbol)

	fmt.Println("✓ Successfully deployed and interacted with ERC20 token")
}
