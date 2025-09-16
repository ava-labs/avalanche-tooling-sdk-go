// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccountServer implements the gRPC AccountService
type AccountServer struct {
	proto.UnimplementedAccountServiceServer
	walletServer *WalletServer // Reference to wallet server for account access
}

// NewAccountServer creates a new AccountServer
func NewAccountServer(walletServer *WalletServer) *AccountServer {
	return &AccountServer{
		walletServer: walletServer,
	}
}

// GetAddresses returns addresses for a specific account
func (s *AccountServer) GetAddresses(ctx context.Context, req *proto.GetAddressesRequest) (*proto.GetAddressesResponse, error) {
	// Get account from wallet server
	acc, exists := s.walletServer.accounts[req.AccountId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "account not found: %s", req.AccountId)
	}

	// Convert addresses to strings
	addresses := make([]string, len(acc.Addresses()))
	for i, addr := range acc.Addresses() {
		addresses[i] = addr.String()
	}

	return &proto.GetAddressesResponse{
		Addresses: addresses,
	}, nil
}

// GetPChainAddress returns the P-Chain address for a specific account and network
func (s *AccountServer) GetPChainAddress(ctx context.Context, req *proto.GetPChainAddressRequest) (*proto.GetPChainAddressResponse, error) {
	// Get account from wallet server
	acc, exists := s.walletServer.accounts[req.AccountId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "account not found: %s", req.AccountId)
	}

	// Parse network
	var net network.Network
	switch req.Network {
	case "fuji":
		net = network.FujiNetwork()
	case "mainnet":
		net = network.MainnetNetwork()
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported network: %s", req.Network)
	}

	// Get P-Chain address
	pChainAddr, err := acc.GetPChainAddress(net)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get P-Chain address: %v", err)
	}

	return &proto.GetPChainAddressResponse{
		PChainAddress: pChainAddr,
	}, nil
}

// GetKeychain returns keychain information for a specific account
func (s *AccountServer) GetKeychain(ctx context.Context, req *proto.GetKeychainRequest) (*proto.GetKeychainResponse, error) {
	// Get account from wallet server
	acc, exists := s.walletServer.accounts[req.AccountId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "account not found: %s", req.AccountId)
	}

	// Get keychain
	keychain, err := acc.GetKeychain()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get keychain: %v", err)
	}

	// Convert addresses to strings
	addresses := make([]string, len(keychain.Addresses().List()))
	for i, addr := range keychain.Addresses().List() {
		addresses[i] = addr.String()
	}

	return &proto.GetKeychainResponse{
		Addresses: addresses,
	}, nil
}
