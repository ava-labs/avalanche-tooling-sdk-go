// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanchego/utils/formatting/address"

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

func (w *LocalWallet) Clients() ChainClients {
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
		return nil, fmt.Errorf("error when importing Account %s \n", err)
	}
	w.AddAccount(existingAccount)
	return &existingAccount, nil
}

// BuildTx constructs a transaction for the specified operation
func (w *LocalWallet) BuildTx(ctx context.Context, params BuildTxParams) (tx.BuildTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network); err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case "P-Chain":
		return w.buildPChainTx(ctx, params.Account, params.BuildTxInput)
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
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network); err != nil {
		return tx.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	if err := w.P().Signer().Sign(context.Background(), params.BuildTxResult.Tx); err != nil {
		return tx.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}
	return tx.SignTxResult{Tx: params.BuildTxResult.Tx}, nil
}

// SendTx submits a signed transaction to the Network
func (w *LocalWallet) SendTx(ctx context.Context, params SendTxParams) (tx.SendTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network); err != nil {
		return tx.SendTxResult{}, fmt.Errorf("error sending tx: %w", err)
	}
	sentTx, err := w.Commit(*params.SignTxResult)
	if err != nil {
		return tx.SendTxResult{}, fmt.Errorf("error sending tx: %w", err)
	}
	return tx.SendTxResult{Tx: sentTx}, nil
}

func (w *LocalWallet) SignPChainTx(ctx context.Context, unsignedTx avagoTxs.UnsignedTx, account account.Account) (*avagoTxs.Tx, error) {
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

func (w *LocalWallet) buildPChainTx(ctx context.Context, account account.Account, params BuildTxInput) (tx.BuildTxResult, error) {
	switch txType := params.GetTxType(); txType {
	case "CreateSubnetTx":
		createSubnetParams, ok := params.(*txs.CreateSubnetTxParams)
		if !ok {
			return tx.BuildTxResult{}, fmt.Errorf("invalid params type for ConvertSubnetToL1Tx, expected *txs.ConvertSubnetToL1TxParams")
		}
		return w.buildCreateSubnetTx(ctx, createSubnetParams)
	case "ConvertSubnetToL1Tx":
		convertParams, ok := params.(*txs.ConvertSubnetToL1TxParams)
		if !ok {
			return tx.BuildTxResult{}, fmt.Errorf("invalid params type for ConvertSubnetToL1Tx, expected *txs.ConvertSubnetToL1TxParams")
		}
		return w.buildConvertSubnetToL1Tx(ctx, account, convertParams)
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

// buildConvertSubnetToL1Tx builds a ConvertSubnetToL1Tx transaction
func (w *LocalWallet) buildCreateSubnetTx(ctx context.Context, params *txs.CreateSubnetTxParams) (tx.BuildTxResult, error) {
	addrs, err := address.ParseToIDs(params.ControlKeys)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("failure parsing control keys: %w", err)
	}
	owners := &secp256k1fx.OutputOwners{
		Addrs:     addrs,
		Threshold: params.Threshold,
		Locktime:  0,
	}
	unsignedTx, err := w.P().Builder().NewCreateSubnetTx(
		owners,
	)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	return tx.BuildTxResult{Tx: &builtTx}, nil
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
	keychain, err := account.GetKeychain()
	if err != nil {
		// Handle error appropriately - for now, return empty options
		return options
	}
	walletAddrs := keychain.Addresses().List()
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

func (w *LocalWallet) Commit(transaction tx.SignTxResult) (*avagoTxs.Tx, error) {
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
