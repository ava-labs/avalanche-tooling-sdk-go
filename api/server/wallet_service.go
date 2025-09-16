// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// WalletServer implements the gRPC WalletService
type WalletServer struct {
	proto.UnimplementedWalletServiceServer
}

// NewWalletServer creates a new WalletServer
func NewWalletServer() (*WalletServer, error) {
	return &WalletServer{}, nil
}

// CreateAccount creates a new account
func (s *WalletServer) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}

// GetAccount retrieves an account by address
func (s *WalletServer) GetAccount(ctx context.Context, req *proto.GetAccountRequest) (*proto.GetAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}

// ListAccounts returns all accounts
func (s *WalletServer) ListAccounts(ctx context.Context, req *emptypb.Empty) (*proto.ListAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}

// ImportAccount imports an existing account
func (s *WalletServer) ImportAccount(ctx context.Context, req *proto.ImportAccountRequest) (*proto.ImportAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportAccount not implemented")
}

// GetAddresses returns all addresses from all accounts
func (s *WalletServer) GetAddresses(ctx context.Context, req *emptypb.Empty) (*proto.GetAddressesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAddresses not implemented")
}

// BuildTransaction builds a transaction
func (s *WalletServer) BuildTransaction(ctx context.Context, req *proto.BuildTransactionRequest) (*proto.BuildTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildTransaction not implemented")
}

// SignTransaction signs a transaction
func (s *WalletServer) SignTransaction(ctx context.Context, req *proto.SignTransactionRequest) (*proto.SignTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignTransaction not implemented")
}

// SendTransaction sends a transaction
func (s *WalletServer) SendTransaction(ctx context.Context, req *proto.SendTransactionRequest) (*proto.SendTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTransaction not implemented")
}

// GetChainClients returns chain client endpoints
func (s *WalletServer) GetChainClients(ctx context.Context, req *emptypb.Empty) (*proto.GetChainClientsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChainClients not implemented")
}

// SetChainClients updates chain client endpoints
func (s *WalletServer) SetChainClients(ctx context.Context, req *proto.SetChainClientsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetChainClients not implemented")
}

// Close performs cleanup
func (s *WalletServer) Close(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
