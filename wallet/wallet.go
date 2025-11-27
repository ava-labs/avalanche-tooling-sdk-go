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

	// SetNetwork sets the active network for wallet operations
	SetNetwork(network network.Network)

	// Network returns the active network for wallet operations
	Network() network.Network

	// =========================================================================
	// Account Management
	// =========================================================================

	// Accounts returns all accounts managed by this wallet
	// Returns map[name]Account for easy lookup by account name
	Accounts() map[string]account.Account

	// ImportAccount imports an account into the wallet
	// The account name is taken from the Account.Name() method and must be unique within the wallet
	ImportAccount(account account.Account) error

	// Account returns a specific account by name
	Account(name string) (account.Account, error)

	// SetActiveAccount sets the active account for operations
	// Automatically set when first adding an account
	SetActiveAccount(name string) error

	// ActiveAccountName returns the currently active account name
	ActiveAccountName() string

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
