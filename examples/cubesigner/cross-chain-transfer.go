//go:build cubesigner_cross_chain_transfer
// +build cubesigner_cross_chain_transfer

// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/session"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/cubesigner"
)

func main() {
	// Get key IDs from environment
	pChainKeyID := os.Getenv("PCHAIN_KEY_ID")
	cChainKeyID := os.Getenv("CCHAIN_KEY_ID")
	if pChainKeyID == "" || cChainKeyID == "" {
		log.Printf("PCHAIN_KEY_ID and CCHAIN_KEY_ID environment variables are required")
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

	// Create keychains
	kcPChain, err := cubesigner.NewKeychain(apiClient, []string{pChainKeyID})
	if err != nil {
		log.Printf("Failed to create P-chain keychain: %v", err)
		return
	}

	kcCChain, err := cubesigner.NewKeychain(apiClient, []string{cChainKeyID})
	if err != nil {
		log.Printf("Failed to create C-chain keychain: %v", err)
		return
	}

	// Create wallet for P-chain operations
	ctx := context.Background()
	walletPChain, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kcPChain,
		kcPChain,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Printf("Failed to create P-chain wallet: %v", err)
		return
	}

	// Get chain and asset IDs
	cChainID := walletPChain.C().Builder().Context().BlockchainID
	avaxAssetID := walletPChain.P().Builder().Context().AVAXAssetID

	// Export 0.5 AVAX from P-chain to C-chain
	exportOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     kcCChain.Addresses().List(),
	}

	exportTx, err := walletPChain.P().IssueExportTx(
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
		log.Printf("Failed to export from P-chain: %v", err)
		return
	}

	fmt.Printf("Exported from P-chain, tx ID: %s\n", exportTx.ID())

	// Wait for export to be processed
	time.Sleep(10 * time.Second)

	// Create wallet for C-chain operations
	walletCChain, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kcCChain,
		kcCChain,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Printf("Failed to create C-chain wallet: %v", err)
		return
	}

	// Import on C-chain
	cChainAddr := kcCChain.EthAddresses().List()[0]
	importTx, err := walletCChain.C().IssueImportTx(
		constants.PlatformChainID,
		cChainAddr,
		common.WithContext(ctx),
	)
	if err != nil {
		log.Printf("Failed to import on C-chain: %v", err)
		return
	}

	fmt.Printf("Imported on C-chain, tx ID: %s\n", importTx.ID())
}
