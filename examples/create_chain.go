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
//   - PRIVATE_KEY: Hex-encoded private key (e.g., "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
//   - SUBNET_AUTH_KEY: P-Chain bech32 address with subnet auth rights (e.g., "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu")
//   - SUBNET_ID: The subnet ID (e.g., "2DeHa7Qb6sufPkmQcFWG2uCd4pBPv9WB6dkzroiMQhd1NSRtof")
//   - EVM_ADDRESS: EVM address for genesis allocation (e.g., "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
//   - BLOCKCHAIN_NAME: Name for the new blockchain (optional, defaults to "MyBlockchain")
func CreateChain() error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()

	// Get required environment variables
	// PRIVATE_KEY should be a hex-encoded private key string (64 characters)
	// Example: "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		return fmt.Errorf("PRIVATE_KEY environment variable is required (hex-encoded private key)")
	}

	// SUBNET_AUTH_KEY should be a P-Chain bech32 address with control over the subnet
	// Example: "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu"
	subnetAuthKey := os.Getenv("SUBNET_AUTH_KEY")
	if subnetAuthKey == "" {
		return fmt.Errorf("SUBNET_AUTH_KEY environment variable is required (P-Chain bech32 address)")
	}

	// SUBNET_ID should be the subnet ID where the blockchain will be created
	// Example: "2DeHa7Qb6sufPkmQcFWG2uCd4pBPv9WB6dkzroiMQhd1NSRtof"
	subnetID := os.Getenv("SUBNET_ID")
	if subnetID == "" {
		return fmt.Errorf("SUBNET_ID environment variable is required")
	}

	// EVM_ADDRESS should be an Ethereum-style address for genesis allocation
	// Example: "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"
	evmAddress := os.Getenv("EVM_ADDRESS")
	if evmAddress == "" {
		return fmt.Errorf("EVM_ADDRESS environment variable is required (EVM address format)")
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
	if err := CreateChain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
