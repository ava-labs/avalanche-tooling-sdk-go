//go:build create_chain
// +build create_chain

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/blockchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// CreateChain demonstrates creating a blockchain on a subnet using the wallet
// Required environment variables:
//   - PRIVATE_KEY: Hex-encoded private key for the account
//   - SUBNET_AUTH_KEY: P-Chain address that has subnet auth rights
//   - SUBNET_ID: The subnet ID where the blockchain will be created
//   - EVM_ADDRESS: EVM address for genesis allocation
//   - BLOCKCHAIN_NAME: Name for the new blockchain
func CreateChain() error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()

	// Get required environment variables
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		return fmt.Errorf("PRIVATE_KEY environment variable is required")
	}

	subnetAuthKey := os.Getenv("SUBNET_AUTH_KEY")
	if subnetAuthKey == "" {
		return fmt.Errorf("SUBNET_AUTH_KEY environment variable is required")
	}

	subnetID := os.Getenv("SUBNET_ID")
	if subnetID == "" {
		return fmt.Errorf("SUBNET_ID environment variable is required")
	}

	evmAddress := os.Getenv("EVM_ADDRESS")
	if evmAddress == "" {
		return fmt.Errorf("EVM_ADDRESS environment variable is required")
	}

	blockchainName := os.Getenv("BLOCKCHAIN_NAME")
	if blockchainName == "" {
		blockchainName = "MyBlockchain"
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

	// Create EVM genesis
	evmGenesisParams := blockchain.GetDefaultSubnetEVMGenesis(evmAddress)
	evmGenesisBytes, err := blockchain.CreateEvmGenesis(&evmGenesisParams)
	if err != nil {
		return fmt.Errorf("failed to create EVM genesis: %w", err)
	}

	// Get VM ID
	vmID, err := blockchain.VMID(blockchainName)
	if err != nil {
		return fmt.Errorf("failed to get VMID: %w", err)
	}

	// Create chain transaction parameters
	createChainParams := &pchainTxs.CreateChainTxParams{
		SubnetAuthKeys: []string{subnetAuthKey},
		SubnetID:       subnetID,
		VMID:           vmID.String(),
		ChainName:      blockchainName,
		Genesis:        evmGenesisBytes,
	}

	// Use SubmitTx to build, sign, and send in one call
	submitTxParams := types.SubmitTxParams{
		BuildTxInput: createChainParams,
	}
	submitTxResult, err := localWallet.Primary().SubmitTx(ctx, submitTxParams)
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
	if err := CreateChain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
