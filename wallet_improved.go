// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package avalanche_tooling_sdk_go

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/transaction"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/network"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/coreth/accounts"
	"github.com/ava-labs/subnet-evm/ethclient"
)

type ChainClients struct {
	C *ethclient.Client // …/ext/bc/C/rpc
	X string            // …/ext/bc/X
	P string            // …/ext/bc/P
}

// LocalWallet represents a local wallet implementation
type LocalWallet struct {
	*primary.Wallet
	accounts []account.Account
	clients  ChainClients
}

// Ensure LocalWallet implements Wallet interface
var _ Wallet = (*LocalWallet)(nil)

// NewLocalWallet creates a new local wallet
func NewLocalWallet(ctx context.Context, uri string) (*LocalWallet, error) {
	wallet, err := primary.MakeWallet(
		ctx,
		uri,
		nil,
		nil,
		primary.WalletConfig{},
	)
	if err != nil {
		return nil, err
	}
	return &LocalWallet{
		Wallet:   wallet,
		accounts: []account.Account{},
		clients:  ChainClients{},
	}, nil
}

func (w *LocalWallet) Accounts() []account.Account {
	return w.accounts
}

func (w *LocalWallet) Clients() ChainClients {
	return w.clients
}

// CreateAccount creates a new account using local key generation
func (w *LocalWallet) CreateAccount(ctx context.Context, network network.Network) (*account.Account, error) {
	newAccount, err := account.NewAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to create new account: %w", err)
	}

	// Add the account to the wallet
	w.AddAccount(newAccount)

	return &newAccount, nil
}

// GetAccount retrieves an existing account by address or identifier
func (w *LocalWallet) GetAccount(ctx context.Context, network network.Network, address ids.ShortID) (*account.Account, error) {
	// TODO: Implement account retrieval logic based on address
	// This could search through w.accounts or use the embedded primary.Wallet
	return nil, fmt.Errorf("not implemented")
}

// ListAccounts returns all accounts managed by this wallet
func (w *LocalWallet) ListAccounts(ctx context.Context, network network.Network) ([]*account.Account, error) {
	// Return all accounts in the wallet
	accounts := w.GetAllAccounts()
	result := make([]*account.Account, len(accounts))
	for i := range accounts {
		result[i] = &accounts[i]
	}
	return result, nil
}

// ImportAccount imports an existing account into the wallet
func (w *LocalWallet) ImportAccount(ctx context.Context, network network.Network, account account.Account) (*account.Account, error) {
	// TODO: Implement account import logic
	// This would add the provided account to the wallet
	return nil, fmt.Errorf("not implemented")
}

// BuildTx constructs a transaction for the specified operation
func (w *LocalWallet) BuildTx(ctx context.Context, network network.Network, txType string, params map[string]interface{}) (transaction.Transaction, error) {
	// TODO: Implement transaction building logic
	// This would use the embedded primary.Wallet to build transactions
	return nil, fmt.Errorf("not implemented")
}

// SignTx signs a transaction
func (w *LocalWallet) SignTx(ctx context.Context, network network.Network, tx transaction.Transaction) (transaction.Transaction, error) {
	// TODO: Implement transaction signing logic
	// This would use the embedded primary.Wallet to sign transactions
	return nil, fmt.Errorf("not implemented")
}

// SendTx submits a signed transaction to the network
func (w *LocalWallet) SendTx(ctx context.Context, network network.Network, tx transaction.Transaction) (ids.ID, error) {
	// TODO: Implement transaction sending logic
	// This would use the embedded primary.Wallet to send transactions
	return ids.Empty, fmt.Errorf("not implemented")
}

func (w *LocalWallet) SignPChainTx(ctx context.Context, unsignedTx txs.UnsignedTx, account accounts.Account) (*txs.Tx, error) {
	tx := txs.Tx{Unsigned: unsignedTx}
	if err := w.Wallet.P().Signer().Sign(ctx, &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return &tx, nil
}

// GetAddresses returns all addresses managed by this wallet
func (w *LocalWallet) GetAddresses(ctx context.Context) ([]ids.ShortID, error) {
	// Get addresses from all accounts in the wallet
	var allAddresses []ids.ShortID
	accounts := w.GetAllAccounts()

	for _, acc := range accounts {
		addresses := acc.Addresses()
		allAddresses = append(allAddresses, addresses...)
	}

	return allAddresses, nil
}

// GetChainClients returns the blockchain clients associated with this wallet
func (w *LocalWallet) GetChainClients() ChainClients {
	return w.clients
}

// SetChainClients updates the blockchain clients for this wallet
func (w *LocalWallet) SetChainClients(clients ChainClients) {
	w.clients = clients
}

// Close performs cleanup operations for the wallet
func (w *LocalWallet) Close(ctx context.Context) error {
	// TODO: Implement cleanup logic if needed
	return nil
}

// AddAccount adds an account to the wallet
func (w *LocalWallet) AddAccount(acc account.Account) {
	w.accounts = append(w.accounts, acc)
}

// GetAccountByAddress finds an account by its address
func (w *LocalWallet) GetAccountByAddress(address ids.ShortID) *account.Account {
	for i := range w.accounts {
		// Check if the account's SoftKey has this address
		addresses := w.accounts[i].Addresses()
		for _, addr := range addresses {
			if addr == address {
				return &w.accounts[i]
			}
		}
	}
	return nil
}

// GetAllAccounts returns all accounts in the wallet
func (w *LocalWallet) GetAllAccounts() []account.Account {
	return w.accounts
}
