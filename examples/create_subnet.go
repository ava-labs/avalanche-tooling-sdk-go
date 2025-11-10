//go:build create_subnet
// +build create_subnet

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

// CreateSubnet demonstrates creating a subnet using the wallet
// Required environment variables:
//   - PRIVATE_KEY: Hex-encoded private key (e.g., "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
//   - CONTROL_KEY_ADDRESS: P-Chain bech32 address for subnet control (e.g., "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu")
func CreateSubnet() error {
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

	// Use SubmitTx to build, sign, and send in one call
	submitTxParams := types.SubmitTxParams{
		BuildTxInput: createSubnetParams,
	}
	submitTxResult, err := localWallet.SubmitTx(ctx, submitTxParams)
	if err != nil {
		return fmt.Errorf("failed to submit tx: %w", err)
	}

	// Print transaction result
	if tx := submitTxResult.GetTx(); tx != nil {
		if pChainTx, ok := tx.(*avagoTxs.Tx); ok {
			fmt.Printf("Transaction ID: %s\n", pChainTx.ID())
		} else {
			fmt.Printf("Transaction: %s\n", submitTxResult.GetChainType())
		}
	}
	return nil
}

func main() {
	if err := CreateSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
