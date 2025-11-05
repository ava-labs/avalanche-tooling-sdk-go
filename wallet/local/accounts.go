// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// generateAccountName generates an automatic account name based on the current count
func (w *LocalWallet) generateAccountName() string {
	return fmt.Sprintf("account-%d", len(w.accounts)+1)
}

// setWalletAccount loads an account into the underlying primary wallet
func (w *LocalWallet) setWalletAccount(ctx context.Context, acc account.Account, net network.Network) error {
	wallet, err := getWalletFromAccount(ctx, acc, net)
	if err != nil {
		return err
	}
	w.wallet = wallet
	return nil
}

// Accounts returns all accounts managed by this wallet with their info
func (w *LocalWallet) Accounts() map[string]account.AccountInfo {
	result := make(map[string]account.AccountInfo)
	for name := range w.accounts {
		info, _ := w.Account(name)
		result[name] = info
	}
	return result
}

// CreateAccount creates a new account with an optional name
func (w *LocalWallet) CreateAccount(name string) (account.AccountInfo, error) {
	// Generate automatic name if empty
	if name == "" {
		name = w.generateAccountName()
	}

	// Check if name already exists
	if _, exists := w.accounts[name]; exists {
		return account.AccountInfo{}, fmt.Errorf("account with name %q already exists", name)
	}

	// Create new local account
	newAccount, err := account.NewLocalAccount()
	if err != nil {
		return account.AccountInfo{}, fmt.Errorf("failed to create new account: %w", err)
	}

	// Store in map
	w.accounts[name] = newAccount

	// Set as active if it's the first account
	if len(w.accounts) == 1 {
		if err := w.SetActiveAccount(name); err != nil {
			return account.AccountInfo{}, fmt.Errorf("failed to set active account: %w", err)
		}
	}

	// Return account info
	return w.Account(name)
}

// Account returns info for a specific account by name
func (w *LocalWallet) Account(name string) (account.AccountInfo, error) {
	acc, exists := w.accounts[name]
	if !exists {
		return account.AccountInfo{}, fmt.Errorf("account %q not found", name)
	}

	pAddr, err := acc.GetPChainAddress(w.defaultNetwork)
	if err != nil {
		return account.AccountInfo{}, err
	}

	xAddr, err := acc.GetXChainAddress(w.defaultNetwork)
	if err != nil {
		return account.AccountInfo{}, err
	}

	cAddr, err := acc.GetCChainAddress(w.defaultNetwork)
	if err != nil {
		return account.AccountInfo{}, err
	}

	evmAddr, err := acc.GetEVMAddress()
	if err != nil {
		return account.AccountInfo{}, err
	}

	return account.AccountInfo{
		Name:       name,
		PAddress:   pAddr,
		XAddress:   xAddr,
		CAddress:   cAddr,
		EVMAddress: evmAddr,
	}, nil
}

// ImportAccount imports an account with a name
func (w *LocalWallet) ImportAccount(name string, spec account.AccountSpec) (account.AccountInfo, error) {
	// Generate automatic name if empty
	if name == "" {
		name = w.generateAccountName()
	}

	// Check if name already exists
	if _, exists := w.accounts[name]; exists {
		return account.AccountInfo{}, fmt.Errorf("account with name %q already exists", name)
	}

	// Import from private key
	if spec.PrivateKey != "" {
		acc, err := account.ImportFromString(spec.PrivateKey)
		if err != nil {
			return account.AccountInfo{}, fmt.Errorf("failed to import account: %w", err)
		}

		// Store in map
		w.accounts[name] = acc

		// Set as active if it's the first account
		if len(w.accounts) == 1 {
			if err := w.SetActiveAccount(name); err != nil {
				return account.AccountInfo{}, fmt.Errorf("failed to set active account: %w", err)
			}
		}

		return w.Account(name)
	}

	// TODO: Handle DerivationPath and RemoteKeyID
	return account.AccountInfo{}, errors.New("only PrivateKey import is currently supported")
}

// ExportAccount exports an account by name
func (w *LocalWallet) ExportAccount(name string) (account.AccountSpec, error) {
	acc, exists := w.accounts[name]
	if !exists {
		return account.AccountSpec{}, fmt.Errorf("account %q not found", name)
	}

	// For local accounts, we can export the private key
	if localAcc, ok := acc.(*account.LocalAccount); ok {
		privKey, err := localAcc.PrivateKey()
		if err != nil {
			return account.AccountSpec{}, fmt.Errorf("failed to export private key: %w", err)
		}
		return account.AccountSpec{
			PrivateKey: privKey,
		}, nil
	}

	// For other account types, return error
	return account.AccountSpec{}, fmt.Errorf("account type does not support export")
}

// SetActiveAccount sets the default account for operations
func (w *LocalWallet) SetActiveAccount(name string) error {
	// Do nothing if already active
	if name == w.activeAccount {
		return nil
	}

	acc, exists := w.accounts[name]
	if !exists {
		return fmt.Errorf("account %q not found", name)
	}
	w.activeAccount = name

	// Load account into the wallet for P/X/C operations
	ctx, cancel := utils.GetWalletRefreshContext()
	defer cancel()
	if err := w.setWalletAccount(ctx, acc, w.defaultNetwork); err != nil {
		return fmt.Errorf("failed to load account into wallet: %w", err)
	}

	return nil
}

// ActiveAccount returns the currently active account name
func (w *LocalWallet) ActiveAccount() string {
	return w.activeAccount
}

// getWalletFromAccount creates an avalanchego wallet from an account
func getWalletFromAccount(ctx context.Context, acc account.Account, net network.Network) (*primary.Wallet, error) {
	keychain, err := acc.GetKeychain()
	if err != nil {
		return nil, err
	}
	wallet, err := primary.MakeWallet(
		ctx,
		net.Endpoint,
		keychain,
		keychain,
		primary.WalletConfig{},
	)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
