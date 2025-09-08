// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package avalanche_tooling_sdk_go

import (
	"context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/transaction"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/network"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/coreth/accounts"
	"github.com/ava-labs/subnet-evm/ethclient"
)

// ChainClientsInterface represents the different blockchain clients available
// This is an interface version of ChainClients to avoid conflicts
type ChainClientsInterface interface {
	GetCChainClient() *ethclient.Client
	GetXChainEndpoint() string
	GetPChainEndpoint() string
}

// WalletConfig represents configuration options for wallet initialization
type WalletConfig struct {
	URI     string
	Network network.Network
	Clients ChainClients
}

// Wallet represents the core wallet interface that can be implemented
// by different wallet types (local, API-based, etc.)
type Wallet interface {
	// Accounts returns the accounts in the Wallet
	Accounts() []account.Account

	// Signer returns the clients in the Wallet
	Clients() ChainClients
	// Account Management
	// CreateAccount creates a new account
	CreateAccount(ctx context.Context, network network.Network) (*account.Account, error)

	// GetAccount retrieves an existing account by address or identifier
	GetAccount(ctx context.Context, network network.Network, address ids.ShortID) (*account.Account, error)

	// ListAccounts returns all accounts managed by this wallet
	ListAccounts(ctx context.Context, network network.Network) ([]*account.Account, error)

	// ImportAccount imports an existing account into the wallet
	ImportAccount(ctx context.Context, network network.Network, account account.Account) (*account.Account, error)

	// Transaction Operations
	// BuildTx constructs a transaction for the specified operation
	BuildTx(ctx context.Context, network network.Network, txType string, params map[string]interface{}) (transaction.Transaction, error)

	// SignTx signs a transaction
	SignTx(ctx context.Context, network network.Network, tx transaction.Transaction) (transaction.Transaction, error)

	// SendTx submits a signed transaction to the network
	SendTx(ctx context.Context, network network.Network, tx transaction.Transaction) (ids.ID, error)

	// P-Chain Specific Operations
	// SignPChainTx signs a P-Chain transaction
	SignPChainTx(ctx context.Context, unsignedTx txs.UnsignedTx, account accounts.Account) (*txs.Tx, error)

	// GetAddresses returns all addresses managed by this wallet
	GetAddresses(ctx context.Context) ([]ids.ShortID, error)

	// GetChainClients returns the blockchain clients associated with this wallet
	GetChainClients() ChainClients

	// SetChainClients updates the blockchain clients for this wallet
	SetChainClients(clients ChainClients)

	// Close performs cleanup operations for the wallet
	Close(ctx context.Context) error
}
