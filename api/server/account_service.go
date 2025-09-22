// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
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

// GetPChainAddress returns the P-Chain address for a specific account and network
func (s *AccountServer) GetPChainAddress(ctx context.Context, req *proto.GetPChainAddressRequest) (*proto.GetPChainAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPChainAddress not implemented")
}

// GetKeychain returns keychain information for a specific account
func (s *AccountServer) GetKeychain(ctx context.Context, req *proto.GetKeychainRequest) (*proto.GetKeychainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKeychain not implemented")
}
