package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create the client
	client := proto.NewWalletServiceClient(conn)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call CreateAccount
	fmt.Println("Calling CreateAccount...")
	req := &proto.CreateAccountRequest{}

	resp, err := client.CreateAccount(ctx, req)
	if err != nil {
		log.Fatalf("CreateAccount failed: %v", err)
	}

	// Print the response
	fmt.Printf("CreateAccount Response:\n")
	fmt.Printf("  FujiAvaxAddress: %s\n", resp.FujiAvaxAddress)
	fmt.Printf("  AvaxAddress: %v\n", resp.AvaxAddress)
	fmt.Printf("  EthAddress: %s\n", resp.EthAddress)

	//Test GetAccount using the Fuji address
	fmt.Println("\nCalling GetAccount...")

	getReq := &proto.GetAccountRequest{
		Address: "fuji1dyy5snc3esyye0z8cq2e39a2lz30mqxvlqk8j2",
	}

	getResp, err := client.GetAccount(ctx, getReq)
	if err != nil {
		log.Fatalf("GetAccount failed: %v", err)
	}

	// Print the GetAccount response
	fmt.Printf("GetAccount Response:\n")
	fmt.Printf("  Address: %s\n", getResp.Address)
	fmt.Printf("  Policies: %v\n", getResp.Policies)
	fmt.Printf("  Created At: %s\n", getResp.CreatedAt)
	fmt.Printf("  Updated At: %s\n", getResp.UpdatedAt)
	fmt.Printf("  Key Type: %s\n", getResp.KeyType)

	fmt.Println("\nTest completed successfully!")
}
