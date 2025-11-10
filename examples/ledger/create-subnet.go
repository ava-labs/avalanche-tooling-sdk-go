//go:build ledger_create_subnet
// +build ledger_create_subnet

// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// create-subnet demonstrates how to create a subnet using a Ledger hardware wallet.
//
// This example shows:
// - Connecting to a Ledger device
// - Creating a keychain from the Ledger
// - Building and signing a CreateSubnet transaction
// - Issuing the transaction to the network
//
// Prerequisites:
// - Ledger device connected and unlocked
// - Avalanche app open on the Ledger
// - Sufficient AVAX balance on Fuji testnet for the Ledger address
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger"
)

func main() {
	// Connect to Ledger device
	// Make sure your Ledger is connected, unlocked, and the Avalanche app is open
	fmt.Println("Connecting to Ledger device...")
	device, err := ledger.New()
	if err != nil {
		log.Fatalf("Failed to connect to Ledger: %v", err)
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			log.Printf("Failed to disconnect from Ledger: %v", err)
		}
	}()

	// Create keychain using the first address (index 0)
	kc, err := ledger.NewKeychain(device, 1)
	if err != nil {
		log.Printf("Failed to create keychain: %v", err)
		return
	}

	// Create wallet on Fuji testnet
	ctx := context.Background()
	wallet, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kc,
		kc,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Printf("Failed to create wallet: %v", err)
		return
	}

	// Create subnet with threshold of 1
	subnetOwnerAddrs := kc.Addresses().List()
	owner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     subnetOwnerAddrs,
	}

	fmt.Println("Building subnet creation transaction...")
	fmt.Println("*** Please confirm the transaction on your Ledger device ***")

	// Sign and issue create subnet transaction
	signCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	createSubnetTx, err := wallet.P().IssueCreateSubnetTx(
		owner,
		common.WithContext(signCtx),
	)
	if err != nil {
		log.Printf("Failed to create subnet: %v", err)
		return
	}

	fmt.Printf("✓ Subnet created with ID: %s\n", createSubnetTx.ID())
}
