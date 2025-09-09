// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	"github.com/ava-labs/avalanchego/ids"
	avagoNetwork "github.com/ava-labs/avalanchego/network"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func main() {
	ctx := context.Background()

	// Create a new local wallet
	wallet, err := NewLocalWallet(ctx, "https://api.avax-test.network")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	// Create an account
	acc, err := wallet.CreateAccount(ctx)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	// Load the account into the wallet
	network := avagoNetwork.Testnet
	err = wallet.loadAccountIntoWallet(ctx, *acc, network)
	if err != nil {
		log.Fatalf("Failed to load account into wallet: %v", err)
	}

	// Example: Build a ConvertSubnetToL1Tx using the new BuildTx method
	convertParams := txs.ConvertSubnetToL1TxParams{
		SubnetAuthKeys: []ids.ShortID{acc.Addresses()[0]},
		SubnetID:       ids.GenerateTestID(),                                 // Replace with actual subnet ID
		ChainID:        ids.GenerateTestID(),                                 // Replace with actual chain ID
		Address:        []byte("0x1234567890abcdef1234567890abcdef12345678"), // Contract address
		Validators: []*txs.ConvertSubnetToL1Validator{
			{
				// Add actual validator details here
				// NodeID: someNodeID,
				// Weight: someWeight,
				// StartTime: someStartTime,
				// EndTime: someEndTime,
			},
		},
		Wallet: wallet.Wallet, // This would be properly initialized
	}

	// Use the new BuildTx method with typed parameters
	tx, err := wallet.BuildTx(ctx, network, convertParams)
	if err != nil {
		log.Fatalf("Failed to build ConvertSubnetToL1Tx: %v", err)
	}

	fmt.Printf("Successfully built ConvertSubnetToL1Tx: %s\n", tx.GetID())
	fmt.Printf("Transaction type: %s\n", tx.GetChainType())
	fmt.Printf("Is signed: %t\n", tx.IsSigned())

	// The transaction is now ready for signing and sending
	// You can proceed with:
	// 1. SignTx() to sign the transaction
	// 2. SendTx() to submit it to the network
}
