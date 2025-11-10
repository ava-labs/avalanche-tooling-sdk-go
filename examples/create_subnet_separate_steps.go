//go:build create_subnet_separate_steps
// +build create_subnet_separate_steps

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// CreateSubnetSeparateSteps demonstrates creating a subnet using separate BuildTx, SignTx, and SendTx calls
// This example shows the individual steps that SubmitTx performs internally.
// Required environment variables:
//   - PRIVATE_KEY: Hex-encoded private key (e.g., "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
//   - CONTROL_KEY_ADDRESS: P-Chain bech32 address for subnet control (e.g., "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu")
func CreateSubnetSeparateSteps() error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()

	// Get required environment variables
	// PRIVATE_KEY should be a hex-encoded private key string (64 characters)
	// Example: "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		return fmt.Errorf("PRIVATE_KEY environment variable is required (hex-encoded private key)")
	}

	// CONTROL_KEY_ADDRESS should be a P-Chain bech32 address
	// Example: "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu"
	controlKeyAddress := os.Getenv("CONTROL_KEY_ADDRESS")
	if controlKeyAddress == "" {
		return fmt.Errorf("CONTROL_KEY_ADDRESS environment variable is required (P-Chain bech32 address format)")
	}

	// Create a local wallet with Fuji network
	net := network.FujiNetwork()
	localWallet, err := local.NewLocalWallet(net)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	// Import account from private key
	accountSpec := account.AccountSpec{
		PrivateKey: privateKey,
	}
	accountInfo, err := localWallet.ImportAccount("my-account", accountSpec)
	if err != nil {
		return fmt.Errorf("failed to import account: %w", err)
	}
	fmt.Printf("Imported account: %s\n", accountInfo.Name)
	fmt.Printf("  P-Chain: %s\n", accountInfo.PAddress)

	// Create subnet transaction parameters
	createSubnetParams := &pchainTxs.CreateSubnetTxParams{
		ControlKeys: []string{controlKeyAddress},
		Threshold:   1,
	}

	// Step 1: Build the transaction
	fmt.Println("\nStep 1: Building transaction...")
	buildTxParams := types.BuildTxParams{
		BuildTxInput: createSubnetParams,
	}
	buildTxResult, err := localWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return fmt.Errorf("failed to build tx: %w", err)
	}
	fmt.Println("✓ Transaction built successfully")

	// Step 2: Sign the transaction
	fmt.Println("\nStep 2: Signing transaction...")
	signTxParams := types.SignTxParams{
		BuildTxResult: &buildTxResult,
	}
	signTxResult, err := localWallet.SignTx(ctx, signTxParams)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}
	fmt.Println("✓ Transaction signed successfully")

	// Step 3: Send the transaction
	fmt.Println("\nStep 3: Sending transaction to network...")
	sendTxParams := types.SendTxParams{
		SignTxResult: &signTxResult,
	}
	sendTxResult, err := localWallet.SendTx(ctx, sendTxParams)
	if err != nil {
		return fmt.Errorf("failed to send tx: %w", err)
	}
	fmt.Println("✓ Transaction sent successfully")

	// Print transaction result
	if tx := sendTxResult.GetTx(); tx != nil {
		if pChainTx, ok := tx.(*avagoTxs.Tx); ok {
			fmt.Printf("\nSubnet created! Transaction ID: %s\n", pChainTx.ID())
		} else {
			fmt.Printf("\nTransaction: %s\n", sendTxResult.GetChainType())
		}
	}
	return nil
}

func main() {
	if err := CreateSubnetSeparateSteps(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
