// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// Accounts returns all accounts managed by this wallet
func (w *LocalWallet) Accounts() map[string]account.Account {
	return maps.Clone(w.accounts)
}

// Account returns a specific account by name
func (w *LocalWallet) Account(name string) (account.Account, error) {
	acc, exists := w.accounts[name]
	if !exists {
		return nil, fmt.Errorf("account %q not found", name)
	}
	return acc, nil
}

// ImportAccount imports an account into the wallet
func (w *LocalWallet) ImportAccount(account account.Account) error {
	name := account.Name()
	if name == "" {
		return errors.New("account name cannot be empty")
	}

	// Check if name already exists
	if _, exists := w.accounts[name]; exists {
		return fmt.Errorf("account with name %q already exists", name)
	}

	// Store in map
	w.accounts[name] = account

	// Set as active if it's the first account
	if len(w.accounts) == 1 {
		if err := w.SetActiveAccount(name); err != nil {
			return fmt.Errorf("failed to set active account: %w", err)
		}
	}

	return nil
}

// SetActiveAccount sets the active account for operations
func (w *LocalWallet) SetActiveAccount(name string) error {
	// Do nothing if already active
	if name == w.activeAccountName {
		return nil
	}

	acc, exists := w.accounts[name]
	if !exists {
		return fmt.Errorf("account %q not found", name)
	}
	w.activeAccountName = name
	w.seenSubnetIDs = []ids.ID{}

	// Load account into the wallet for P/X/C operations
	ctx, cancel := utils.GetPrimaryWalletCreationContext()
	defer cancel()
	wallet, err := createPrimaryWallet(ctx, acc, w.activeNetwork, nil)
	if err != nil {
		return fmt.Errorf("failed to load account into wallet: %w", err)
	}
	w.primaryWallet = wallet

	return nil
}

// ActiveAccountName returns the currently active account name
func (w *LocalWallet) ActiveAccountName() string {
	return w.activeAccountName
}

// createPrimaryWallet creates an avalanchego wallet from an account
func createPrimaryWallet(ctx context.Context, acc account.Account, network network.Network, walletConfig *primary.WalletConfig) (*primary.Wallet, error) {
	keychain, err := acc.GetKeychain()
	if err != nil {
		return nil, err
	}

	config := primary.WalletConfig{}
	if walletConfig != nil {
		config = *walletConfig
	}

	wallet, err := primary.MakeWallet(
		ctx,
		network.Endpoint,
		keychain,
		keychain,
		config,
	)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
