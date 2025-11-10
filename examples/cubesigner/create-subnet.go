//go:build cubesigner_create_subnet
// +build cubesigner_create_subnet

// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/session"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/cubesigner"
)

func main() {
	// Get key ID from environment
	keyID := os.Getenv("CUBESIGNER_KEY_ID")
	if keyID == "" {
		log.Printf("CUBESIGNER_KEY_ID environment variable is required")
		return
	}

	// Initialize CubeSigner client
	sessionFile := "session.json"
	manager, err := session.NewJsonSessionManager(&sessionFile)
	if err != nil {
		log.Printf("Failed to create session manager: %v", err)
		return
	}

	apiClient, err := client.NewApiClient(manager)
	if err != nil {
		log.Printf("Failed to create API client: %v", err)
		return
	}

	// Create CubeSigner keychain
	kc, err := cubesigner.NewKeychain(apiClient, []string{keyID})
	if err != nil {
		log.Printf("Failed to create keychain: %v", err)
		return
	}

	// Create primary wallet on Fuji testnet
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

	// Issue create subnet transaction
	createSubnetTx, err := wallet.P().IssueCreateSubnetTx(
		owner,
		common.WithContext(ctx),
	)
	if err != nil {
		log.Printf("Failed to create subnet: %v", err)
		return
	}

	fmt.Printf("Subnet created with ID: %s\n", createSubnetTx.ID())
}
