// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create clients for both services
	walletClient := proto.NewWalletServiceClient(conn)
	accountClient := proto.NewAccountServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Avalanche Tooling SDK Server API Example ===")
	fmt.Println()

	// 1. Create a new account
	fmt.Println("1. Creating a new account...")
	createResp, err := walletClient.CreateAccount(ctx, &proto.CreateAccountRequest{})
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	fmt.Printf("✅ Account created successfully!\n")
	fmt.Printf("   Fuji AVAX Address: %s\n", createResp.FujiAvaxAddress)
	fmt.Printf("   Mainnet AVAX Address: %s\n", createResp.AvaxAddress)
	fmt.Printf("   ETH Address: %s\n", createResp.EthAddress)
	fmt.Println()

	// 2. Try to get account (currently unimplemented)
	fmt.Println("2. Attempting to get account...")
	getResp, err := walletClient.GetAccount(ctx, &proto.GetAccountRequest{
		Address: createResp.FujiAvaxAddress,
	})
	if err != nil {
		fmt.Printf("❌ GetAccount failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ Account retrieved: %v\n", getResp)
	}
	fmt.Println()

	// 3. Try to list accounts (currently unimplemented)
	fmt.Println("3. Attempting to list accounts...")
	listResp, err := walletClient.ListAccounts(ctx, &emptypb.Empty{})
	if err != nil {
		fmt.Printf("❌ ListAccounts failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ Accounts listed: %v\n", listResp)
	}
	fmt.Println()

	// 4. Try to get P-Chain address (currently unimplemented)
	fmt.Println("4. Attempting to get P-Chain address...")
	pChainResp, err := accountClient.GetPChainAddress(ctx, &proto.GetPChainAddressRequest{
		AccountId: createResp.FujiAvaxAddress,
		Network:   "fuji",
	})
	if err != nil {
		fmt.Printf("❌ GetPChainAddress failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ P-Chain address: %s\n", pChainResp.PChainAddress)
	}
	fmt.Println()

	// 5. Try to get keychain (currently unimplemented)
	fmt.Println("5. Attempting to get keychain...")
	keychainResp, err := accountClient.GetKeychain(ctx, &proto.GetKeychainRequest{
		AccountId: createResp.FujiAvaxAddress,
	})
	if err != nil {
		fmt.Printf("❌ GetKeychain failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ Keychain: %v\n", keychainResp)
	}
	fmt.Println()

	// 6. Try to build a transaction (currently unimplemented)
	fmt.Println("6. Attempting to build a transaction...")
	buildTxResp, err := walletClient.BuildTransaction(ctx, &proto.BuildTransactionRequest{
		AccountId: createResp.FujiAvaxAddress,
		Network:   "fuji",
		TransactionParams: &proto.TransactionParams{
			TxType:    "create_subnet",
			ChainType: "p_chain",
			PChainParams: &proto.PChainParams{
				Params: &proto.PChainParams_CreateSubnet{
					CreateSubnet: &proto.CreateSubnetParams{
						ControlKeys: []string{createResp.FujiAvaxAddress},
						Threshold:   1,
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("❌ BuildTransaction failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ Transaction built: %s\n", buildTxResp.TransactionId)
	}
	fmt.Println()

	// 7. Try to get chain clients (currently unimplemented)
	fmt.Println("7. Attempting to get chain clients...")
	chainClientsResp, err := walletClient.GetChainClients(ctx, &emptypb.Empty{})
	if err != nil {
		fmt.Printf("❌ GetChainClients failed (expected - not implemented): %v\n", err)
	} else {
		fmt.Printf("✅ Chain clients: C-Chain=%s, X-Chain=%s, P-Chain=%s\n",
			chainClientsResp.CChainEndpoint,
			chainClientsResp.XChainEndpoint,
			chainClientsResp.PChainEndpoint)
	}
	fmt.Println()

	fmt.Println("=== Example completed! ===")
	fmt.Println("Note: Most methods are currently unimplemented.")
	fmt.Println("Only CreateAccount is fully functional at this time.")
}
