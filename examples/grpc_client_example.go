// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create API wallet that connects to gRPC server
	apiWallet, err := wallet.NewAPIWallet("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create API wallet: %v", err)
	}
	defer apiWallet.Close(ctx)

	fmt.Println("Connected to gRPC server!")

	// Create a new account
	fmt.Println("Creating new account...")
	account, err := apiWallet.CreateAccount(ctx)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	fmt.Printf("Created account with addresses: %v\n", account.Addresses())

	// List all accounts
	fmt.Println("Listing all accounts...")
	accounts, err := apiWallet.ListAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}

	fmt.Printf("Found %d accounts:\n", len(accounts))
	for i, acc := range accounts {
		fmt.Printf("  Account %d: %v\n", i+1, acc.Addresses())
	}

	// Get P-Chain address
	fmt.Println("Getting P-Chain address...")
	pChainAddr, err := account.GetPChainAddress(network.FujiNetwork())
	if err != nil {
		log.Printf("Failed to get P-Chain address: %v", err)
	} else {
		fmt.Printf("P-Chain address: %s\n", pChainAddr)
	}

	// Example: Build a CreateSubnet transaction
	fmt.Println("Building CreateSubnet transaction...")
	createSubnetParams := &txs.CreateSubnetTxParams{
		ControlKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		Threshold:   1,
	}

	buildTxParams := wallet.BuildTxParams{
		BuildTxInput: createSubnetParams,
		Account:      *account,
		Network:      network.FujiNetwork(),
	}

	buildTxResult, err := apiWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		log.Printf("Failed to build transaction: %v", err)
	} else {
		fmt.Printf("Successfully built transaction: %v\n", buildTxResult)
	}

	// Get all addresses
	fmt.Println("Getting all addresses...")
	addresses, err := apiWallet.GetAddresses(ctx)
	if err != nil {
		log.Printf("Failed to get addresses: %v", err)
	} else {
		fmt.Printf("All addresses: %v\n", addresses)
	}

	fmt.Println("Example completed successfully!")
}
