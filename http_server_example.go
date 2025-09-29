package main

import (
	"log"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/server"
)

func main() {
	// Create and start the combined server (gRPC + HTTP)
	srv, err := server.NewCombinedServer(":8080", ":8081")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server started:")
	log.Println("  gRPC server: localhost:8080")
	log.Println("  HTTP server: localhost:8081")
	log.Println("  HTTP endpoints:")
	log.Println("    POST /v1/wallet/accounts - Create account")
	log.Println("    GET  /v1/wallet/accounts/{address} - Get account")
	log.Println("    GET  /v1/wallet/accounts - List accounts")
	log.Println("    POST /v1/wallet/accounts/import - Import account")
	log.Println("    POST /v1/wallet/transactions/build - Build transaction")
	log.Println("    POST /v1/wallet/transactions/sign - Sign transaction")
	log.Println("    POST /v1/wallet/transactions/send - Send transaction")

	// Wait for shutdown signal
	srv.WaitForShutdown()
}
