package main

import (
	"log"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/server"
)

func main() {
	// Create and start the gRPC server
	srv, err := server.NewServer("8080")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for shutdown signal
	srv.WaitForShutdown()
}

