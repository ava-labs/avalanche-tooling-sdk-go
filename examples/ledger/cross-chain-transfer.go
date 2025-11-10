//go:build ledger_cross_chain_transfer
// +build ledger_cross_chain_transfer

// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// cross-chain-transfer demonstrates how to transfer AVAX from P-Chain to C-Chain using a Ledger.
//
// This example shows:
// - Exporting AVAX from P-Chain
// - Importing AVAX on C-Chain
// - Deriving Ethereum address from Ledger public key
//
// Prerequisites:
// - Ledger device connected and unlocked
// - Avalanche app open on the Ledger
// - Sufficient AVAX balance on Fuji P-Chain for the Ledger address
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger"
)

func main() {
	// Connect to Ledger
	device, err := ledger.New()
	if err != nil {
		log.Fatalf("Failed to connect to Ledger: %v", err)
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			log.Printf("Failed to disconnect from Ledger: %v", err)
		}
	}()

	// Create keychain
	kc, err := ledger.NewKeychain(device, 1)
	if err != nil {
		log.Printf("Failed to create keychain: %v", err)
		return
	}

	// Create wallet
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

	// Get chain and asset IDs
	cChainID := wallet.C().Builder().Context().BlockchainID
	avaxAssetID := wallet.P().Builder().Context().AVAXAssetID

	// Export 0.5 AVAX from P-Chain to C-Chain
	exportOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     kc.Addresses().List(),
	}

	fmt.Println("Exporting from P-Chain...")
	fmt.Println("*** Please confirm the transaction on your Ledger device ***")

	exportTx, err := wallet.P().IssueExportTx(
		cChainID,
		[]*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt:          units.Avax / 2, // 0.5 AVAX
				OutputOwners: *exportOwner,
			},
		}},
		common.WithContext(ctx),
	)
	if err != nil {
		log.Printf("Failed to export from P-Chain: %v", err)
		return
	}

	fmt.Printf("✓ Exported from P-Chain, tx ID: %s\n", exportTx.ID())

	// Wait for export to be processed
	time.Sleep(10 * time.Second)

	// Import on C-Chain
	// Get Ethereum address from the Ledger-derived public key
	pubKey, err := device.PubKey(0)
	if err != nil {
		log.Printf("Failed to get public key: %v", err)
		return
	}
	cChainAddr := pubKey.EthAddress()

	fmt.Println("Importing on C-Chain...")
	fmt.Println("*** Please confirm the transaction on your Ledger device ***")

	importTx, err := wallet.C().IssueImportTx(
		constants.PlatformChainID,
		cChainAddr,
		common.WithContext(ctx),
	)
	if err != nil {
		log.Printf("Failed to import on C-Chain: %v", err)
		return
	}

	fmt.Printf("✓ Imported on C-Chain, tx ID: %s\n", importTx.ID())
}
