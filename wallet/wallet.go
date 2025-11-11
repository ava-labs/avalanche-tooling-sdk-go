// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// ChainClients is now defined in wallet/types/common.go

// Wallet represents the core wallet interface that can be implemented
// by different wallet types (local, API-based, etc.)
type Wallet interface {
	// =========================================================================
	// Network Management
	// =========================================================================

	// SetNetwork sets the default network for wallet operations
	SetNetwork(net network.Network)

	// Network returns the default network for wallet operations
	Network() network.Network

	// =========================================================================
	// Account Management
	// =========================================================================

	// Accounts returns all accounts managed by this wallet with their info
	// Returns map[name]AccountInfo for easy lookup by account name
	Accounts() map[string]account.AccountInfo

	// CreateAccount creates a new account with an optional name.
	// If name is empty, generates a default name (e.g., "account-1")
	// Returns the account info including all chain addresses.
	CreateAccount(name string) (account.AccountInfo, error)

	// ImportAccount imports an account with a name
	// Returns the imported account info
	ImportAccount(name string, spec account.AccountSpec) (account.AccountInfo, error)

	// ExportAccount exports an account by name
	// WARNING: For local accounts, this exposes the private key!
	ExportAccount(name string) (account.AccountSpec, error)

	// Account returns info for a specific account by name
	Account(name string) (account.AccountInfo, error)

	// SetActiveAccount sets the default account for operations
	// Automatically set when first adding an account
	SetActiveAccount(name string) error

	// ActiveAccount returns the currently active account name
	ActiveAccount() string

	// =========================================================================
	// Transaction Operations
	// =========================================================================

	// BuildTx constructs a transaction for the specified operation
	BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error)

	// SignTx signs a transaction
	SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error)

	// SendTx submits a signed transaction to the Network
	SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error)
}
