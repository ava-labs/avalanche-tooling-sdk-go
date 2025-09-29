// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// CombinedServer represents a server that runs both gRPC and HTTP
type CombinedServer struct {
	grpcServer   *grpc.Server
	httpServer   *HTTPServer
	grpcAddr     string
	httpAddr     string
	walletServer *WalletServer
}

// NewCombinedServer creates a new combined server
func NewCombinedServer(grpcAddr, httpAddr string) (*CombinedServer, error) {
	// Create wallet server
	walletServer, err := NewWalletServer()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet server: %w", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	proto.RegisterWalletServiceServer(grpcServer, walletServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Create HTTP server
	httpServer, err := NewHTTPServer(grpcAddr, httpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	return &CombinedServer{
		grpcServer:   grpcServer,
		httpServer:   httpServer,
		grpcAddr:     grpcAddr,
		httpAddr:     httpAddr,
		walletServer: walletServer,
	}, nil
}

// Start starts both gRPC and HTTP servers
func (s *CombinedServer) Start() error {
	// Start gRPC server
	lis, err := net.Listen("tcp", s.grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %s: %w", s.grpcAddr, err)
	}

	log.Printf("Starting gRPC server on %s", s.grpcAddr)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Wait a moment for gRPC server to start
	time.Sleep(100 * time.Millisecond)

	// Start HTTP server
	if err := s.httpServer.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop gracefully stops both servers
func (s *CombinedServer) Stop() error {
	log.Println("Stopping combined server...")

	// Stop HTTP server first
	if err := s.httpServer.Stop(); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}

	// Stop gRPC server
	s.grpcServer.GracefulStop()

	// Wallet cleanup (if needed)
	// Note: WalletServer doesn't have a Close method

	log.Println("Combined server stopped")
	return nil
}

// WaitForShutdown waits for shutdown signal
func (s *CombinedServer) WaitForShutdown() {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Stop server
	if err := s.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
}

// GetWalletServer returns the wallet server instance
func (s *CombinedServer) GetWalletServer() *WalletServer {
	return s.walletServer
}
