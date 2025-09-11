// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/ids"
)

type BuildTxParams struct {
	BuildTxInput
	Account account.Account
	Network network.Network
}

// BuildTxInput represents a generic interface for transaction parameters
type BuildTxInput interface {
	// GetTxType returns the transaction type identifier
	GetTxType() string
	// Validate validates the parameters
	Validate() error
	// GetChainType returns which chain this transaction is for
	GetChainType() string
}

type SignTxParams struct {
	*tx.BuildTxResult
	Account account.Account
	Network network.Network
}

type SendTxParams struct {
	*tx.SignTxResult
	Account account.Account
	Network network.Network
}

// Wallet represents the core wallet interface that can be implemented
// by different wallet types (local, API-based, etc.)
type Wallet interface {
	// Accounts returns the accounts in the Wallet
	Accounts() []account.Account

	// Signer returns the clients in the Wallet
	Clients() ChainClients
	// Account Management
	// CreateAccount creates a new Account
	CreateAccount(ctx context.Context) (*account.Account, error)

	// GetAccount retrieves an existing Account by address or identifier
	GetAccount(ctx context.Context, address ids.ShortID) (*account.Account, error)

	// ListAccounts returns all accounts managed by this wallet
	ListAccounts(ctx context.Context) ([]*account.Account, error)

	// ImportAccount imports an existing Account into the wallet
	ImportAccount(ctx context.Context, keyPath string) (*account.Account, error)

	// Transaction Operations
	// BuildTx constructs a transaction for the specified operation
	BuildTx(ctx context.Context, params BuildTxParams) (tx.BuildTxResult, error)

	// SignTx signs a transaction
	SignTx(ctx context.Context, params SignTxParams) (tx.SignTxResult, error)

	// SendTx submits a signed transaction to the Network
	SendTx(ctx context.Context, params SendTxParams) (tx.SendTxResult, error)

	// GetAddresses returns all addresses managed by this wallet
	GetAddresses(ctx context.Context) ([]ids.ShortID, error)

	// GetChainClients returns the blockchain clients associated with this wallet
	GetChainClients() ChainClients

	// SetChainClients updates the blockchain clients for this wallet
	SetChainClients(clients ChainClients)

	// Close performs cleanup operations for the wallet
	Close(ctx context.Context) error
}
