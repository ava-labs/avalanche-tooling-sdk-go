// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	"github.com/ava-labs/avalanchego/ids"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
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
func NewLocalWallet() (*LocalWallet, error) {
	return &LocalWallet{
		Wallet:   nil,
		accounts: []account.Account{},
		clients:  ChainClients{},
	}, nil
}

func (w *LocalWallet) loadAccountIntoWallet(ctx context.Context, account account.Account, network network.Network) error {
	wallet, err := primary.MakeWallet(
		ctx,
		network.Endpoint,
		account.SoftKey.KeyChain(),
		account.SoftKey.KeyChain(),
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

func (w *LocalWallet) Clients() ChainClients {
	return w.clients
}

// CreateAccount creates a new account using local key generation
func (w *LocalWallet) CreateAccount(ctx context.Context) (*account.Account, error) {
	newAccount, err := account.NewAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to create new account: %w", err)
	}

	// Add the account to the wallet
	w.AddAccount(newAccount)

	return &newAccount, nil
}

// GetAccount retrieves an existing account by address or identifier
func (w *LocalWallet) GetAccount(ctx context.Context, address ids.ShortID) (*account.Account, error) {
	// TODO: Implement account retrieval logic based on address
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

// ImportAccount imports an existing account into the wallet
func (w *LocalWallet) ImportAccount(ctx context.Context, keyPath string) (*account.Account, error) {
	// TODO: Implement account import logic
	// This would add the provided account to the wallet
	existingAccount, err := account.Import(keyPath)
	if err != nil {
		return nil, fmt.Errorf("error when importing account %s \n", err)
	}
	w.AddAccount(existingAccount)
	return &existingAccount, nil
}

// BuildTx constructs a transaction for the specified operation
func (w *LocalWallet) BuildTx(ctx context.Context, params BuildTxParams) (tx.BuildTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.account, params.network); err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case "P-Chain":
		return w.buildPChainTx(ctx, params.account, params.BuildTxInput)
	case "C-Chain":
		return w.buildCChainTx(ctx, params)
	case "X-Chain":
		return w.buildXChainTx(ctx, params)
	default:
		return tx.BuildTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

// SignTx signs a transaction
func (w *LocalWallet) SignTx(ctx context.Context, params SignTxParams) (tx.SignTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.account, params.network); err != nil {
		return tx.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	if err := w.P().Signer().Sign(context.Background(), params.BuildTxResult.Tx); err != nil {
		return tx.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	return tx.SignTxResult{Tx: params.BuildTxResult.Tx}, nil
}

// SendTx submits a signed transaction to the network
func (w *LocalWallet) SendTx(ctx context.Context, params SendTxParams) (tx.SendTxResult, error) {
	// TODO: Implement transaction sending logic
	// This would use the embedded primary.Wallet to send transactions
	return tx.SendTxResult{}, fmt.Errorf("not implemented")
}

func (w *LocalWallet) SignPChainTx(ctx context.Context, unsignedTx avagoTxs.UnsignedTx, account accounts.Account) (*avagoTxs.Tx, error) {
	tx := avagoTxs.Tx{Unsigned: unsignedTx}
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

func (w *LocalWallet) buildPChainTx(ctx context.Context, account account.Account, params BuildTxInput) (tx.BuildTxResult, error) {
	switch txType := params.GetTxType(); txType {
	case "ConvertSubnetToL1Tx":
		convertParams, ok := params.(*txs.ConvertSubnetToL1TxParams)
		if !ok {
			return tx.BuildTxResult{}, fmt.Errorf("invalid params type for ConvertSubnetToL1Tx, expected *txs.ConvertSubnetToL1TxParams")
		}
		return w.buildConvertSubnetToL1Tx(ctx, account, convertParams)
	case "DisableL1ValidatorTx":
		disableParams, ok := params.(*txs.DisableL1ValidatorTxParams)
		if !ok {
			return tx.BuildTxResult{}, fmt.Errorf("invalid params type for DisableL1ValidatorTx, expected *txs.DisableL1ValidatorTxParams")
		}
		return w.buildDisableL1ValidatorTx(ctx, disableParams)
	default:
		return tx.BuildTxResult{}, fmt.Errorf("unsupported P-Chain transaction type: %s", txType)
	}
}

// buildConvertSubnetToL1Tx builds a ConvertSubnetToL1Tx transaction
func (w *LocalWallet) buildConvertSubnetToL1Tx(ctx context.Context, account account.Account, params *txs.ConvertSubnetToL1TxParams) (tx.BuildTxResult, error) {
	options := GetMultisigTxOptions(account, params.SubnetAuthKeys)
	unsignedTx, err := w.P().Builder().NewConvertSubnetToL1Tx(
		params.SubnetID,
		params.ChainID,
		params.Address,
		params.Validators,
		options...,
	)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	return tx.BuildTxResult{Tx: &builtTx}, nil
}

// buildDisableL1ValidatorTx builds a DisableL1ValidatorTx transaction
func (w *LocalWallet) buildDisableL1ValidatorTx(ctx context.Context, params *txs.DisableL1ValidatorTxParams) (tx.BuildTxResult, error) {
	//// Use the existing NewDisableL1ValidatorTx function
	//signedTx, err := txs.NewDisableL1ValidatorTx(*params)
	//if err != nil {
	//	return tx.BuildTxResult{}, fmt.Errorf("failed to create DisableL1ValidatorTx: %w", err)
	//}
	//
	//// Convert SignTxResult to BuildTxResult
	//return tx.BuildTxResult{
	//	Tx: signedTx.Tx,
	//}, nil
	return tx.BuildTxResult{
		Tx: nil,
	}, nil
}

// buildCChainTx builds C-Chain transactions
func (w *LocalWallet) buildCChainTx(ctx context.Context, params BuildTxInput) (tx.BuildTxResult, error) {
	// TODO: Implement C-Chain transaction building
	return tx.BuildTxResult{}, fmt.Errorf("C-Chain transactions not yet implemented")
}

// buildXChainTx builds X-Chain transactions
func (w *LocalWallet) buildXChainTx(ctx context.Context, params BuildTxInput) (tx.BuildTxResult, error) {
	// TODO: Implement X-Chain transaction building
	return tx.BuildTxResult{}, fmt.Errorf("X-Chain transactions not yet implemented")
}

func GetMultisigTxOptions(account account.Account, subnetAuthKeys []ids.ShortID) []common.Option {
	options := []common.Option{}
	walletAddrs := account.SoftKey.KeyChain().Addresses().List()
	changeAddr := walletAddrs[0]
	// addrs to use for signing
	customAddrsSet := set.Set[ids.ShortID]{}
	customAddrsSet.Add(walletAddrs...)
	customAddrsSet.Add(subnetAuthKeys...)
	options = append(options, common.WithCustomAddresses(customAddrsSet))
	// set change to go to wallet addr (instead of any other subnet auth key)
	changeOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{changeAddr},
	}
	options = append(options, common.WithChangeOwner(changeOwner))
	return options
}
