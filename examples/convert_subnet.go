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

	"github.com/ava-labs/avalanchego/utils/logging"

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
//   - PRIVATE_KEY: Hex-encoded private key for the account
//   - SUBNET_AUTH_KEY: P-Chain address that has subnet auth rights
//   - SUBNET_ID: The subnet ID to convert
//   - CHAIN_ID: Blockchain ID of the L1 where validator manager contract is deployed
//   - NODE_ID: NodeID of the bootstrap validator (e.g., NodeID-xxxxx)
//   - BLS_PUBLIC_KEY: BLS public key for the validator
//   - BLS_PROOF_OF_POSSESSION: BLS proof of possession
//   - CHANGE_OWNER_ADDRESS: P-Chain address for remaining balance owner
//   - VALIDATOR_MANAGER_ADDRESS: Address of validator manager contract (optional, defaults to 0x0FEEDC0DE0000000000000000000000000000000)
//   - VALIDATOR_WEIGHT: Weight for the validator (optional, defaults to 100)
//   - VALIDATOR_BALANCE: Balance in nAVAX for the validator (optional, defaults to 1000000000)
func ConvertSubnet() error {
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

	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		return fmt.Errorf("CHAIN_ID environment variable is required")
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		return fmt.Errorf("NODE_ID environment variable is required")
	}

	blsPublicKey := os.Getenv("BLS_PUBLIC_KEY")
	if blsPublicKey == "" {
		return fmt.Errorf("BLS_PUBLIC_KEY environment variable is required")
	}

	blsProofOfPossession := os.Getenv("BLS_PROOF_OF_POSSESSION")
	if blsProofOfPossession == "" {
		return fmt.Errorf("BLS_PROOF_OF_POSSESSION environment variable is required")
	}

	changeOwnerAddress := os.Getenv("CHANGE_OWNER_ADDRESS")
	if changeOwnerAddress == "" {
		return fmt.Errorf("CHANGE_OWNER_ADDRESS environment variable is required")
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
	localWallet, err := local.NewLocalWallet(logging.NoLog{}, net)
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
	if err := ConvertSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
