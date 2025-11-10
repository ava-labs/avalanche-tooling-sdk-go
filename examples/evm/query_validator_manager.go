//go:build query_validator_manager
// +build query_validator_manager

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/validatormanager"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
)

const dexalotTestnetRPC = "https://subnets.avax.network/dexalot/testnet/rpc"

// This example demonstrates how to query validator manager using ReadContract
//
// Environment variables:
//   - RPC_URL: The blockchain RPC URL (defaults to Dexalot testnet if not set)
func main() {
	// Get RPC URL from environment variable
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = dexalotTestnetRPC
	}

	fmt.Printf("Querying validator manager for: %s\n\n", rpcURL)

	// Get validator manager info
	net := network.FujiNetwork()
	vmInfo, err := validatormanager.GetValidatorManagerInfo(net, rpcURL, "")
	if err != nil {
		fmt.Printf("Failed to get validator manager info: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Validator Manager RPC: %s\n", vmInfo.ManagerRPC)
	fmt.Printf("Validator Manager Address: %s\n\n", vmInfo.ManagerAddress)

	// Create wallet and set chain
	w, err := local.NewLocalWallet(logging.NoLog{}, net)
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if err := w.SetChain(vmInfo.ManagerRPC); err != nil {
		fmt.Printf("Failed to set chain: %s\n", err)
		os.Exit(1)
	}

	// Query subnetID
	fmt.Println("=== Calling subnetID() ===")
	method := wallet.Method("subnetID()->(bytes32)")
	result, err := w.ReadContract(vmInfo.ManagerAddress, method)
	if err != nil {
		fmt.Printf("Failed to call subnetID: %s\n", err)
		os.Exit(1)
	}
	subnetIDBytes, err := wallet.ParseResult[[32]byte](result)
	if err != nil {
		fmt.Printf("Failed to parse subnetID: %s\n", err)
		os.Exit(1)
	}
	subnetID, err := ids.ToID(subnetIDBytes[:])
	if err != nil {
		fmt.Printf("Failed to convert subnetID: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Subnet ID: %s\n\n", subnetID)

	// Query l1TotalWeight
	fmt.Println("=== Calling l1TotalWeight() ===")
	method = wallet.Method("l1TotalWeight()->(uint64)")
	result, err = w.ReadContract(vmInfo.ManagerAddress, method)
	if err != nil {
		fmt.Printf("Failed to call l1TotalWeight: %s\n", err)
		os.Exit(1)
	}
	totalWeight, err := wallet.ParseResult[uint64](result)
	if err != nil {
		fmt.Printf("Failed to parse l1TotalWeight: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("L1 Total Weight: %d\n\n", totalWeight)

	// Query getChurnTracker - returns (uint64, uint8, uint256, uint64, uint64, uint64)
	// The struct is flattened: churnPeriodSeconds, maximumChurnPercentage, startTime, initialWeight, totalWeight, churnAmount
	fmt.Println("=== Calling getChurnTracker() ===")
	method = wallet.Method("getChurnTracker()->(uint64,uint8,uint256,uint64,uint64,uint64)")
	result, err = w.ReadContract(vmInfo.ManagerAddress, method)
	if err != nil {
		fmt.Printf("Failed to call getChurnTracker: %s\n", err)
		os.Exit(1)
	}
	churnPeriodSeconds, err := wallet.ParseResult[uint64](result, 0)
	if err != nil {
		fmt.Printf("Failed to parse churnPeriodSeconds: %s\n", err)
		os.Exit(1)
	}
	maximumChurnPercentage, err := wallet.ParseResult[uint8](result, 1)
	if err != nil {
		fmt.Printf("Failed to parse maximumChurnPercentage: %s\n", err)
		os.Exit(1)
	}
	startTime, err := wallet.ParseResult[*big.Int](result, 2)
	if err != nil {
		fmt.Printf("Failed to parse startTime: %s\n", err)
		os.Exit(1)
	}
	initialWeight, err := wallet.ParseResult[uint64](result, 3)
	if err != nil {
		fmt.Printf("Failed to parse initialWeight: %s\n", err)
		os.Exit(1)
	}
	totalWeightChurn, err := wallet.ParseResult[uint64](result, 4)
	if err != nil {
		fmt.Printf("Failed to parse totalWeight: %s\n", err)
		os.Exit(1)
	}
	churnAmount, err := wallet.ParseResult[uint64](result, 5)
	if err != nil {
		fmt.Printf("Failed to parse churnAmount: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Churn Period Seconds: %d\n", churnPeriodSeconds)
	fmt.Printf("Maximum Churn Percentage: %d%%\n", maximumChurnPercentage)
	fmt.Printf("Churn Tracker:\n")
	fmt.Printf("  Start Time: %s\n", startTime.String())
	fmt.Printf("  Initial Weight: %d\n", initialWeight)
	fmt.Printf("  Total Weight: %d\n", totalWeightChurn)
	fmt.Printf("  Churn Amount: %d\n\n", churnAmount)

	fmt.Println("✓ Successfully queried validator manager")
}
