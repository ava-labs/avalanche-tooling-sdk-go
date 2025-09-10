// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	"github.com/ava-labs/avalanchego/ids"
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
	network := network.Testnet
	err = wallet.loadAccountIntoWallet(ctx, *acc, network)
	if err != nil {
		log.Fatalf("Failed to load account into wallet: %v", err)
	}

	// Example 1: Build a ConvertSubnetToL1Tx
	convertParams := txs.ConvertSubnetToL1TxParams{
		SubnetAuthKeys: []ids.ShortID{acc.Addresses()[0]},
		SubnetID:       ids.GenerateTestID(),
		ChainID:        ids.GenerateTestID(),
		Address:        []byte("0x1234567890abcdef"),
		Validators: []*txs.ConvertSubnetToL1Validator{
			{
				// Add validator details here
			},
		},
		Wallet: wallet.Wallet, // This would need to be properly initialized
	}

	tx, err := wallet.BuildTx(ctx, network, convertParams)
	if err != nil {
		log.Fatalf("Failed to build ConvertSubnetToL1Tx: %v", err)
	}

	fmt.Printf("Built transaction: %s\n", tx.GetID())

	// Example 2: Build a DisableL1ValidatorTx
	disableParams := txs.DisableL1ValidatorTxParams{
		ValidationID: ids.GenerateTestID(),
		Wallet:       wallet.Wallet, // This would need to be properly initialized
	}

	tx2, err := wallet.BuildTx(ctx, network, disableParams)
	if err != nil {
		log.Fatalf("Failed to build DisableL1ValidatorTx: %v", err)
	}

	fmt.Printf("Built transaction: %s\n", tx2.GetID())
}
