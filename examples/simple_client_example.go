// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Simple Client Example ===")
	fmt.Println("This example uses the wallet.APIWallet wrapper")
	fmt.Println()

	// Create API wallet that connects to gRPC server
	fmt.Println("Connecting to gRPC server at localhost:8080...")
	apiWallet, err := wallet.NewAPIWallet("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create API wallet: %v", err)
	}
	defer apiWallet.Close(ctx)

	fmt.Println("✅ Connected to server!")
	fmt.Println()

	// Create a new account
	fmt.Println("Creating new account...")
	account, err := apiWallet.CreateAccount(ctx)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	fmt.Printf("✅ Account created successfully!\n")
	fmt.Printf("   Addresses: %v\n", (*account).Addresses())
	fmt.Println()

	// Try to list accounts (this will fail since it's not implemented)
	fmt.Println("Attempting to list accounts...")
	accounts, err := apiWallet.ListAccounts(ctx)
	if err != nil {
		fmt.Printf("❌ ListAccounts failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ Found %d accounts:\n", len(accounts))
		for i, acc := range accounts {
			fmt.Printf("   Account %d: %v\n", i+1, (*acc).Addresses())
		}
	}
	fmt.Println()

	// Try to get all addresses (this will fail since it's not implemented)
	fmt.Println("Attempting to get all addresses...")
	addresses, err := apiWallet.GetAddresses(ctx)
	if err != nil {
		fmt.Printf("❌ GetAddresses failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ All addresses: %v\n", addresses)
	}
	fmt.Println()

	fmt.Println("=== Example completed! ===")
	fmt.Println("Note: Most methods are currently unimplemented.")
	fmt.Println("Only CreateAccount is fully functional at this time.")
}
