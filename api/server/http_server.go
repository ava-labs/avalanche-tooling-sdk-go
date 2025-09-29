// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
)

// HTTPServer represents the HTTP server with gRPC gateway
type HTTPServer struct {
	httpServer   *http.Server
	grpcAddr     string
	httpAddr     string
	walletServer *WalletServer
}

// NewHTTPServer creates a new HTTP server with gRPC gateway
func NewHTTPServer(grpcAddr, httpAddr string) (*HTTPServer, error) {
	// Create wallet server
	walletServer, err := NewWalletServer()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet server: %w", err)
	}

	// Create gRPC connection for gateway
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Create gRPC gateway mux
	gwMux := runtime.NewServeMux()

	// Register services with the gateway
	if err := proto.RegisterWalletServiceHandler(context.Background(), gwMux, conn); err != nil {
		return nil, fmt.Errorf("failed to register wallet service handler: %w", err)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      gwMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return &HTTPServer{
		httpServer:   httpServer,
		grpcAddr:     grpcAddr,
		httpAddr:     httpAddr,
		walletServer: walletServer,
	}, nil
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	log.Printf("Starting HTTP server on %s (proxying to gRPC at %s)", s.httpAddr, s.grpcAddr)

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop() error {
	log.Println("Stopping HTTP server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	log.Println("HTTP server stopped")
	return nil
}

// GetWalletServer returns the wallet server instance
func (s *HTTPServer) GetWalletServer() *WalletServer {
	return s.walletServer
}
