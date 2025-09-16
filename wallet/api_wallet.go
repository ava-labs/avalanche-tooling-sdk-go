// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"
	"github.com/ava-labs/avalanchego/ids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// APIWallet represents a wallet that communicates with a gRPC server
type APIWallet struct {
	grpcClient proto.WalletServiceClient
	conn       *grpc.ClientConn
	accounts   map[string]*account.APIAccount // account_id -> account mapping
}

// NewAPIWallet creates a new API wallet that connects to a gRPC server
func NewAPIWallet(serverAddr string) (*APIWallet, error) {
	// Connect to gRPC server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Create gRPC client
	grpcClient := proto.NewWalletServiceClient(conn)

	return &APIWallet{
		grpcClient: grpcClient,
		conn:       conn,
		accounts:   make(map[string]*account.APIAccount),
	}, nil
}

// Close closes the gRPC connection
func (w *APIWallet) Close(ctx context.Context) error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// Ensure APIWallet implements Wallet interface
var _ Wallet = (*APIWallet)(nil)

// Accounts returns all accounts in the wallet
func (w *APIWallet) Accounts() []account.Account {
	accounts := make([]account.Account, 0, len(w.accounts))
	for _, acc := range w.accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}

// Clients returns chain clients (not implemented for API wallet)
func (w *APIWallet) Clients() ChainClients {
	// For API wallet, we don't maintain local chain clients
	// The server handles chain client management
	return ChainClients{}
}

// CreateAccount creates a new account via the gRPC server
func (w *APIWallet) CreateAccount(ctx context.Context) (*account.Account, error) {
	// Call gRPC server
	resp, err := w.grpcClient.CreateAccount(ctx, &proto.CreateAccountRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Create API account
	apiAccount := &account.APIAccount{
		AccountID:              resp.AccountId,
		ServerAccountAddresses: make([]ids.ShortID, len(resp.Addresses)),
		GrpcClient:             w.grpcClient,
	}

	// Convert addresses
	for i, addrStr := range resp.Addresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid address from server: %s", addrStr)
		}
		apiAccount.ServerAccountAddresses[i] = addr
	}

	// Store account
	w.accounts[resp.AccountId] = apiAccount

	// Return as account.Account interface
	var accountInterface account.Account = apiAccount
	return &accountInterface, nil
}

// GetAccount retrieves an account by address
func (w *APIWallet) GetAccount(ctx context.Context, address ids.ShortID) (*account.Account, error) {
	// Call gRPC server
	resp, err := w.grpcClient.GetAccount(ctx, &proto.GetAccountRequest{
		Address: address.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Check if we already have this account cached
	if apiAccount, exists := w.accounts[resp.AccountId]; exists {
		var accountInterface account.Account = apiAccount
		return &accountInterface, nil
	}

	// Create new API account
	apiAccount := &account.APIAccount{
		AccountID:              resp.AccountId,
		ServerAccountAddresses: make([]ids.ShortID, len(resp.Addresses)),
		GrpcClient:             w.grpcClient,
	}

	// Convert addresses
	for i, addrStr := range resp.Addresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid address from server: %s", addrStr)
		}
		apiAccount.ServerAccountAddresses[i] = addr
	}

	// Store account
	w.accounts[resp.AccountId] = apiAccount

	// Return as account.Account interface
	var accountInterface account.Account = apiAccount
	return &accountInterface, nil
}

// ListAccounts returns all accounts managed by this wallet
func (w *APIWallet) ListAccounts(ctx context.Context) ([]*account.Account, error) {
	// Call gRPC server
	resp, err := w.grpcClient.ListAccounts(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	// Convert to account.Account interfaces
	accounts := make([]*account.Account, 0, len(resp.Accounts))
	for _, accInfo := range resp.Accounts {
		// Check if we already have this account cached
		if apiAccount, exists := w.accounts[accInfo.AccountId]; exists {
			var accountInterface account.Account = apiAccount
			accounts = append(accounts, &accountInterface)
			continue
		}

		// Create new API account
		apiAccount := &account.APIAccount{
			AccountID:              accInfo.AccountId,
			ServerAccountAddresses: make([]ids.ShortID, len(accInfo.Addresses)),
			GrpcClient:             w.grpcClient,
		}

		// Convert addresses
		for i, addrStr := range accInfo.Addresses {
			addr, err := ids.ShortFromString(addrStr)
			if err != nil {
				return nil, fmt.Errorf("invalid address from server: %s", addrStr)
			}
			apiAccount.ServerAccountAddresses[i] = addr
		}

		// Store account
		w.accounts[accInfo.AccountId] = apiAccount

		// Add to result
		var accountInterface account.Account = apiAccount
		accounts = append(accounts, &accountInterface)
	}

	return accounts, nil
}

// ImportAccount imports an existing account
func (w *APIWallet) ImportAccount(ctx context.Context, keyPath string) (*account.Account, error) {
	// Call gRPC server
	resp, err := w.grpcClient.ImportAccount(ctx, &proto.ImportAccountRequest{
		KeyPath: keyPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to import account: %w", err)
	}

	// Create API account
	apiAccount := &account.APIAccount{
		AccountID:              resp.AccountId,
		ServerAccountAddresses: make([]ids.ShortID, len(resp.Addresses)),
		GrpcClient:             w.grpcClient,
	}

	// Convert addresses
	for i, addrStr := range resp.Addresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid address from server: %s", addrStr)
		}
		apiAccount.ServerAccountAddresses[i] = addr
	}

	// Store account
	w.accounts[resp.AccountId] = apiAccount

	// Return as account.Account interface
	var accountInterface account.Account = apiAccount
	return &accountInterface, nil
}

// BuildTx constructs a transaction via the gRPC server
func (w *APIWallet) BuildTx(ctx context.Context, params BuildTxParams) (tx.BuildTxResult, error) {
	// Find account ID
	var accountID string
	for id, acc := range w.accounts {
		for _, addr := range acc.Addresses() {
			for _, paramAddr := range params.Account.Addresses() {
				if addr == paramAddr {
					accountID = id
					break
				}
			}
			if accountID != "" {
				break
			}
		}
		if accountID != "" {
			break
		}
	}

	if accountID == "" {
		return tx.BuildTxResult{}, fmt.Errorf("account not found in wallet")
	}

	// Convert network to string
	var networkStr string
	switch params.Network {
	case network.FujiNetwork():
		networkStr = "fuji"
	case network.MainnetNetwork():
		networkStr = "mainnet"
	default:
		return tx.BuildTxResult{}, fmt.Errorf("unsupported network")
	}

	// Convert transaction params
	txParams, err := w.convertBuildTxInput(params.BuildTxInput)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("failed to convert transaction params: %w", err)
	}

	// Call gRPC server
	_, err = w.grpcClient.BuildTransaction(ctx, &proto.BuildTransactionRequest{
		AccountId:         accountID,
		Network:           networkStr,
		TransactionParams: txParams,
	})
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Convert response to BuildTxResult
	// Note: This is simplified - in a real implementation, you'd need to
	// properly deserialize the transaction based on its type
	return tx.BuildTxResult{
		// This would need proper implementation based on your tx package
	}, nil
}

// SignTx signs a transaction via the gRPC server
func (w *APIWallet) SignTx(ctx context.Context, params SignTxParams) (tx.SignTxResult, error) {
	// Similar to BuildTx, this would call the gRPC server
	return tx.SignTxResult{}, fmt.Errorf("SignTx not fully implemented for API wallet")
}

// SendTx sends a transaction via the gRPC server
func (w *APIWallet) SendTx(ctx context.Context, params SendTxParams) (tx.SendTxResult, error) {
	// Similar to BuildTx, this would call the gRPC server
	return tx.SendTxResult{}, fmt.Errorf("SendTx not fully implemented for API wallet")
}

// GetAddresses returns all addresses managed by this wallet
func (w *APIWallet) GetAddresses(ctx context.Context) ([]ids.ShortID, error) {
	// Call gRPC server
	resp, err := w.grpcClient.GetAddresses(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	// Convert addresses
	addresses := make([]ids.ShortID, len(resp.Addresses))
	for i, addrStr := range resp.Addresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid address from server: %s", addrStr)
		}
		addresses[i] = addr
	}

	return addresses, nil
}

// GetChainClients returns chain clients (not implemented for API wallet)
func (w *APIWallet) GetChainClients() ChainClients {
	// For API wallet, we don't maintain local chain clients
	return ChainClients{}
}

// SetChainClients updates chain clients (not implemented for API wallet)
func (w *APIWallet) SetChainClients(clients ChainClients) {
	// For API wallet, chain clients are managed by the server
}

// convertBuildTxInput converts BuildTxInput to protobuf TransactionParams
func (w *APIWallet) convertBuildTxInput(input BuildTxInput) (*proto.TransactionParams, error) {
	// This is a simplified conversion - you'd need to implement
	// proper conversion based on your transaction types
	return &proto.TransactionParams{
		TxType:    input.GetTxType(),
		ChainType: input.GetChainType(),
		// You'd need to populate the specific chain params based on the input type
	}, nil
}
