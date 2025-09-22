// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents the gRPC server
type Server struct {
	grpcServer    *grpc.Server
	walletServer  *WalletServer
	accountServer *AccountServer
	port          string
}

// NewServer creates a new gRPC server
func NewServer(port string) (*Server, error) {
	// Create wallet server
	walletServer, err := NewWalletServer()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet server: %w", err)
	}

	// Create account server
	accountServer := NewAccountServer(walletServer)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	proto.RegisterWalletServiceServer(grpcServer, walletServer)
	proto.RegisterAccountServiceServer(grpcServer, accountServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	return &Server{
		grpcServer:    grpcServer,
		walletServer:  walletServer,
		accountServer: accountServer,
		port:          port,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	log.Printf("Starting gRPC server on port %s", s.port)

	// Start server in goroutine
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() error {
	log.Println("Stopping gRPC server...")

	// Close wallet
	ctx := context.Background()
	_, err := s.walletServer.Close(ctx, nil)
	if err != nil {
		log.Printf("Error closing wallet: %v", err)
	}

	// Stop gRPC server
	s.grpcServer.GracefulStop()

	log.Println("gRPC server stopped")
	return nil
}

// WaitForShutdown waits for shutdown signal
func (s *Server) WaitForShutdown() {
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
func (s *Server) GetWalletServer() *WalletServer {
	return s.walletServer
}

// GetAccountServer returns the account server instance
func (s *Server) GetAccountServer() *AccountServer {
	return s.accountServer
}
