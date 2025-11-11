//go:build convert_subnet
// +build convert_subnet

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// ConvertSubnet demonstrates converting a subnet to L1 using the wallet
// Required environment variables:
//   - PRIVATE_KEY: Hex-encoded private key (e.g., "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
//   - SUBNET_AUTH_KEY: P-Chain bech32 address with subnet auth rights (e.g., "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu")
//   - SUBNET_ID: The subnet ID to convert (e.g., "2DeHa7Qb6sufPkmQcFWG2uCd4pBPv9WB6dkzroiMQhd1NSRtof")
//   - CHAIN_ID: Blockchain ID where validator manager contract is deployed (e.g., "2CA7jYxRCd5K3qS23JhPybaMsMHSsLc3vGTJxSFDvnBSNdPELz")
//   - NODE_ID: NodeID of the bootstrap validator (e.g., "NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu")
//   - BLS_PUBLIC_KEY: BLS public key for the validator (e.g., "0x80b7851ce335cee149b7cfffbf6cf0bbca3c9b25026a24056e610976d095906e833a66d5ca5c56c23a3fe50e8785a81f")
//   - BLS_PROOF_OF_POSSESSION: BLS proof of possession (e.g., "0x89e1d6d47ff04ec0c78501a029865140e9ec12baba75a95bfc5710b3fecb8db4b6cecb5ccb1136e19f88db0539deb4420306dd60145024197b41cf89179790f20146fba398bc4d13e08540ea812207f736ca007275e4ebdb840065fdb38573de")
//   - CHANGE_OWNER_ADDRESS: P-Chain bech32 address for remaining balance owner (e.g., "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu")
//   - VALIDATOR_MANAGER_ADDRESS: EVM address of validator manager contract (optional, defaults to "0x0FEEDC0DE0000000000000000000000000000000")
//   - VALIDATOR_WEIGHT: Weight for the validator (optional, defaults to 100)
//   - VALIDATOR_BALANCE: Balance in nAVAX for the validator (optional, defaults to 1000000000)
func ConvertSubnet() error {
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

	// SUBNET_ID should be the subnet ID to convert
	// Example: "2DeHa7Qb6sufPkmQcFWG2uCd4pBPv9WB6dkzroiMQhd1NSRtof"
	subnetID := os.Getenv("SUBNET_ID")
	if subnetID == "" {
		return fmt.Errorf("SUBNET_ID environment variable is required")
	}

	// CHAIN_ID should be the blockchain ID where validator manager is deployed
	// Example: "2CA7jYxRCd5K3qS23JhPybaMsMHSsLc3vGTJxSFDvnBSNdPELz"
	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		return fmt.Errorf("CHAIN_ID environment variable is required")
	}

	// NODE_ID should be the NodeID of the bootstrap validator
	// Example: "NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu"
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		return fmt.Errorf("NODE_ID environment variable is required")
	}

	// BLS_PUBLIC_KEY should be a hex-encoded BLS public key
	// Example: "0x80b7851ce335cee149b7cfffbf6cf0bbca3c9b25026a24056e610976d095906e833a66d5ca5c56c23a3fe50e8785a81f"
	blsPublicKey := os.Getenv("BLS_PUBLIC_KEY")
	if blsPublicKey == "" {
		return fmt.Errorf("BLS_PUBLIC_KEY environment variable is required")
	}

	// BLS_PROOF_OF_POSSESSION should be a hex-encoded BLS proof
	// Example: "0x89e1d6d47ff04ec0c78501a029865140e9ec12baba75a95bfc5710b3fecb8db4b6cecb5ccb1136e19f88db0539deb4420306dd60145024197b41cf89179790f20146fba398bc4d13e08540ea812207f736ca007275e4ebdb840065fdb38573de"
	blsProofOfPossession := os.Getenv("BLS_PROOF_OF_POSSESSION")
	if blsProofOfPossession == "" {
		return fmt.Errorf("BLS_PROOF_OF_POSSESSION environment variable is required")
	}

	// CHANGE_OWNER_ADDRESS should be a P-Chain bech32 address for remaining balance owner
	// Example: "P-fuji1zwch24mn3sjkahds98fjd0asudjk2e4ajduu"
	changeOwnerAddress := os.Getenv("CHANGE_OWNER_ADDRESS")
	if changeOwnerAddress == "" {
		return fmt.Errorf("CHANGE_OWNER_ADDRESS environment variable is required (P-Chain bech32 address)")
	}

	// Optional parameters with defaults
	validatorManagerAddress := os.Getenv("VALIDATOR_MANAGER_ADDRESS")
	if validatorManagerAddress == "" {
		validatorManagerAddress = "0x0FEEDC0DE0000000000000000000000000000000"
	}

	validatorWeight := uint64(100)
	if weightStr := os.Getenv("VALIDATOR_WEIGHT"); weightStr != "" {
		weight, err := strconv.ParseUint(weightStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid VALIDATOR_WEIGHT: %w", err)
		}
		validatorWeight = weight
	}

	validatorBalance := uint64(1000000000)
	if balanceStr := os.Getenv("VALIDATOR_BALANCE"); balanceStr != "" {
		balance, err := strconv.ParseUint(balanceStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid VALIDATOR_BALANCE: %w", err)
		}
		validatorBalance = balance
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

	// Create bootstrap validator
	bootstrapValidator := &pchainTxs.ConvertSubnetToL1Validator{
		NodeID:                nodeID,
		Weight:                validatorWeight,
		Balance:               validatorBalance,
		BLSPublicKey:          blsPublicKey,
		BLSProofOfPossession:  blsProofOfPossession,
		RemainingBalanceOwner: changeOwnerAddress,
	}

	// Convert subnet transaction parameters
	convertSubnetParams := &pchainTxs.ConvertSubnetToL1TxParams{
		SubnetAuthKeys: []string{subnetAuthKey},
		SubnetID:       subnetID,
		ChainID:        chainID,
		Validators:     []*pchainTxs.ConvertSubnetToL1Validator{bootstrapValidator},
		Address:        validatorManagerAddress,
	}

	// Use SubmitTx to build, sign, and send in one call
	submitTxParams := types.SubmitTxParams{
		BuildTxInput: convertSubnetParams,
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
	if err := ConvertSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
