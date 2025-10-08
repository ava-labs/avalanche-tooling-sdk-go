// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// ChainClients is now defined in the wallet package

// LocalWallet represents a local wallet implementation
type LocalWallet struct {
	*primary.Wallet
	accounts []account.Account
	clients  types.ChainClients
}

// Ensure LocalWallet implements Wallet interface
var _ wallet.Wallet = (*LocalWallet)(nil)

// NewLocalWallet creates a new local wallet
func NewLocalWallet() (*LocalWallet, error) {
	return &LocalWallet{
		Wallet:   nil,
		accounts: []account.Account{},
		clients:  types.ChainClients{},
	}, nil
}

func (w *LocalWallet) loadAccountIntoWallet(ctx context.Context, account account.Account, network network.Network) error {
	keychain, err := account.GetKeychain()
	if err != nil {
		return err
	}
	wallet, err := primary.MakeWallet(
		ctx,
		network.Endpoint,
		keychain,
		keychain,
		primary.WalletConfig{},
	)
	if err != nil {
		return err
	}
	w.Wallet = wallet
	return nil
}

func (w *LocalWallet) Accounts() []account.Account {
	return w.accounts
}

func (w *LocalWallet) Clients() types.ChainClients {
	return w.clients
}

// CreateAccount creates a new Account using local key generation
func (w *LocalWallet) CreateAccount(ctx context.Context) (*account.Account, error) {
	newAccount, err := account.NewLocalAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to create new Account: %w", err)
	}

	// Add the Account to the wallet
	w.AddAccount(newAccount)

	return &newAccount, nil
}

// GetAccount retrieves an existing Account by address or identifier
func (w *LocalWallet) GetAccount(ctx context.Context, address ids.ShortID) (*account.Account, error) {
	// TODO: Implement Account retrieval logic based on address
	// This could search through w.accounts or use the embedded primary.Wallet
	return nil, fmt.Errorf("not implemented")
}

// ListAccounts returns all accounts managed by this wallet
func (w *LocalWallet) ListAccounts(ctx context.Context) ([]*account.Account, error) {
	// Return all accounts in the wallet
	accounts := w.GetAllAccounts()
	result := make([]*account.Account, len(accounts))
	for i := range accounts {
		result[i] = &accounts[i]
	}
	return result, nil
}

// ImportAccount imports an existing Account into the wallet
func (w *LocalWallet) ImportAccount(ctx context.Context, keyPath string) (*account.Account, error) {
	// TODO: Implement Account import logic
	// This would add the provided Account to the wallet
	existingAccount, err := account.Import(keyPath)
	if err != nil {
		return nil, fmt.Errorf("error when importing Account %w \n", err)
	}
	w.AddAccount(existingAccount)
	return &existingAccount, nil
}

// BuildTx constructs a transaction for the specified operation
func (w *LocalWallet) BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network); err != nil {
		return types.BuildTxResult{}, fmt.Errorf("error loading account into wallet: %w", err)
	}
	return wallet.BuildTx(ctx, w.Wallet, params)
}

// SignTx signs a transaction
func (w *LocalWallet) SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network); err != nil {
		return types.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}

	return wallet.SignTx(ctx, w.Wallet, params)
}

// SendTx submits a signed transaction to the Network
func (w *LocalWallet) SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network); err != nil {
		return types.SendTxResult{}, fmt.Errorf("error loading account into wallet: %w", err)
	}

	return wallet.SendTx(ctx, w.Wallet, params)
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
func (w *LocalWallet) GetChainClients() types.ChainClients {
	return w.clients
}

// SetChainClients updates the blockchain clients for this wallet
func (w *LocalWallet) SetChainClients(clients types.ChainClients) {
	w.clients = clients
}

// Close performs cleanup operations for the wallet
func (w *LocalWallet) Close(ctx context.Context) error {
	// TODO: Implement cleanup logic if needed
	return nil
}

// AddAccount adds an Account to the wallet
func (w *LocalWallet) AddAccount(acc account.Account) {
	w.accounts = append(w.accounts, acc)
}

// GetAccountByAddress finds an Account by its address
func (w *LocalWallet) GetAccountByAddress(address ids.ShortID) *account.Account {
	for i := range w.accounts {
		// Check if the Account's SoftKey has this address
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

func (w *LocalWallet) Commit(transaction types.SignTxResult) (*avagoTxs.Tx, error) {
	if transaction.Undefined() {
		return nil, multisig.ErrUndefinedTx
	}
	isReady, err := transaction.IsReadyToCommit()
	if err != nil {
		return nil, err
	}
	if !isReady {
		return nil, errors.New("tx is not fully signed so can't be committed")
	}
	tx, err := transaction.GetWrappedPChainTx()
	if err != nil {
		return nil, err
	}
	const (
		repeats             = 3
		sleepBetweenRepeats = 2 * time.Second
	)
	var issueTxErr error
	if err != nil {
		return nil, err
	}
	for i := 0; i < repeats; i++ {
		ctx, cancel := utils.GetAPILargeContext()
		defer cancel()
		options := []common.Option{common.WithContext(ctx)}
		// TODO: split error checking and recovery between issuing and waiting for status
		issueTxErr = w.P().IssueTx(tx, options...)
		if issueTxErr == nil {
			break
		}
		if ctx.Err() != nil {
			issueTxErr = fmt.Errorf("timeout issuing/verifying tx with ID %s: %w", tx.ID(), issueTxErr)
		} else {
			issueTxErr = fmt.Errorf("error issuing tx with ID %s: %w", tx.ID(), issueTxErr)
		}
		time.Sleep(sleepBetweenRepeats)
	}
	if issueTxErr != nil {
		return nil, fmt.Errorf("issue tx error %w", issueTxErr)
	}
	return tx, issueTxErr
}
