// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// ChainClients is now defined in wallet/types/common.go

// Wallet represents the core wallet interface that can be implemented
// by different wallet types (local, API-based, etc.)
type Wallet interface {
	// Accounts returns the accounts in the Wallet
	Accounts() []account.Account

	// Signer returns the clients in the Wallet
	Clients() types.ChainClients
	// Account Management
	// CreateAccount creates a new Account
	CreateAccount() (*account.Account, error)

	// GetAccount retrieves an existing Account by address or identifier
	GetAccount(address ids.ShortID) (*account.Account, error)

	// ListAccounts returns all accounts managed by this wallet
	ListAccounts() ([]*account.Account, error)

	// ImportAccount imports an existing Account into the wallet
	ImportAccount(keyPath string) (*account.Account, error)

	// Transaction Operations
	// BuildTx constructs a transaction for the specified operation
	BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error)

	// SignTx signs a transaction
	SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error)

	// SendTx submits a signed transaction to the Network
	SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error)

	// GetAddresses returns all addresses managed by this wallet
	GetAddresses() ([]ids.ShortID, error)

	// GetChainClients returns the blockchain clients associated with this wallet
	GetChainClients() types.ChainClients

	// SetChainClients updates the blockchain clients for this wallet
	SetChainClients(clients types.ChainClients)

	// Close performs cleanup operations for the wallet
	Close() error
}
