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

	txs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// ChainClients is now defined in the wallet package

// LocalWallet represents a local wallet implementation
type LocalWallet struct {
	*primary.Wallet
	accounts []account.Account
}

// Ensure LocalWallet implements Wallet interface
var _ wallet.Wallet = (*LocalWallet)(nil)

// NewLocalWallet creates a new local wallet
func NewLocalWallet() (*LocalWallet, error) {
	return &LocalWallet{
		Wallet:   nil,
		accounts: []account.Account{},
	}, nil
}

// buildWalletConfig creates a WalletConfig with appropriate SubnetIDs based on transaction type
func buildWalletConfig(txInput interface{}) primary.WalletConfig {
	config := primary.WalletConfig{}

	// Handle different input types
	switch input := txInput.(type) {
	case *types.SignTxParams:
		// Extract subnet ID from SignTxParams by looking at the BuildTxResult
		if input != nil && input.BuildTxResult != nil {
			if subnetID, err := extractSubnetIDFromBuildTxResult(input.BuildTxResult); err == nil && subnetID != ids.Empty {
				config.SubnetIDs = []ids.ID{subnetID}
			}
		}
	case *types.SendTxParams:
		// Extract subnet ID from SendTxParams by looking at the SignTxResult
		if input != nil && input.SignTxResult != nil {
			if subnetID, err := extractSubnetIDFromSignTxResult(input.SignTxResult); err == nil && subnetID != ids.Empty {
				config.SubnetIDs = []ids.ID{subnetID}
			}
		}
	case *types.BuildTxParams:
		// Extract subnet ID from BuildTxParams by looking at the BuildTxInput
		if input != nil && input.BuildTxInput != nil {
			if subnetID, err := extractSubnetIDFromBuildTxInput(input.BuildTxInput); err == nil && subnetID != ids.Empty {
				config.SubnetIDs = []ids.ID{subnetID}
			}
		}
	}

	return config
}

// extractSubnetIDFromBuildTxInput extracts subnet ID from BuildTxInput parameters
// This function handles the case where we have transaction parameters but not the built transaction yet
func extractSubnetIDFromBuildTxInput(input types.BuildTxInput) (ids.ID, error) {
	// Handle different BuildTxInput types
	switch params := input.(type) {
	case *txs.CreateChainTxParams:
		// For CreateChainTx, extract subnet ID from parameters
		if params.SubnetID != "" {
			return ids.FromString(params.SubnetID)
		}
	case *txs.ConvertSubnetToL1TxParams:
		// For ConvertSubnetToL1Tx, extract subnet ID from parameters
		if params.SubnetID != "" {
			return ids.FromString(params.SubnetID)
		}
	case *txs.CreateSubnetTxParams:
		// CreateSubnetTx doesn't have a subnet ID since it creates the subnet
		return ids.Empty, fmt.Errorf("CreateSubnetTx doesn't have a subnet ID")
	default:
		// Unknown BuildTxInput type
	}
	return ids.Empty, fmt.Errorf("no subnet ID found in BuildTxInput")
}

// extractSubnetIDFromBuildTxResult extracts subnet ID from a BuildTxResult
func extractSubnetIDFromBuildTxResult(result *types.BuildTxResult) (ids.ID, error) {
	if result != nil && result.BuildTxOutput != nil {
		if tx := result.BuildTxOutput.GetTx(); tx != nil {
			return extractSubnetIDFromTx(tx)
		}
	}
	return ids.Empty, fmt.Errorf("no transaction found in BuildTxResult")
}

// extractSubnetIDFromSignTxResult extracts subnet ID from a SignTxResult
func extractSubnetIDFromSignTxResult(result *types.SignTxResult) (ids.ID, error) {
	if result != nil && result.SignTxOutput != nil {
		if tx := result.SignTxOutput.GetTx(); tx != nil {
			return extractSubnetIDFromTx(tx)
		}
	}
	return ids.Empty, fmt.Errorf("no transaction found in SignTxResult")
}

// extractSubnetIDFromTx extracts subnet ID from a transaction object
func extractSubnetIDFromTx(tx interface{}) (ids.ID, error) {
	// Handle P-Chain transactions
	if pChainTx, ok := tx.(*avagoTxs.Tx); ok && pChainTx.Unsigned != nil {
		switch unsignedTx := pChainTx.Unsigned.(type) {
		case *avagoTxs.CreateChainTx:
			// For CreateChainTx, the subnet ID field is SubnetID
			return unsignedTx.SubnetID, nil
		case *avagoTxs.ConvertSubnetToL1Tx:
			// For CreateChainTx, the subnet ID field is SubnetID
			return unsignedTx.Subnet, nil
		}
	}
	return ids.Empty, fmt.Errorf("no subnet ID found in transaction")
}

func (w *LocalWallet) loadAccountIntoWallet(ctx context.Context, account account.Account, network network.Network, txInput interface{}) error {
	keychain, err := account.GetKeychain()
	if err != nil {
		return err
	}

	walletConfig := buildWalletConfig(txInput)
	wallet, err := primary.MakeWallet(
		ctx,
		network.Endpoint,
		keychain,
		keychain,
		walletConfig,
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

// CreateAccount creates a new Account using local key generation
func (w *LocalWallet) CreateAccount() (*account.Account, error) {
	newAccount, err := account.NewLocalAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to create new Account: %w", err)
	}

	// Add the Account to the wallet
	w.AddAccount(newAccount)

	return &newAccount, nil
}

// ListAccounts returns all accounts managed by this wallet
func (w *LocalWallet) ListAccounts() ([]*account.Account, error) {
	// Return all accounts in the wallet
	accounts := w.GetAllAccounts()
	result := make([]*account.Account, len(accounts))
	for i := range accounts {
		result[i] = &accounts[i]
	}
	return result, nil
}

// ImportAccount imports an existing Account into the wallet
func (w *LocalWallet) ImportAccount(keyPath string) (*account.Account, error) {
	// TODO: Implement Account import logic
	// This would add the provided Account to the wallet
	existingAccount, err := account.Import(keyPath)
	if err != nil {
		return nil, fmt.Errorf("error when importing Account %w", err)
	}
	w.AddAccount(existingAccount)
	return &existingAccount, nil
}

// BuildTx constructs a transaction for the specified operation
func (w *LocalWallet) BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network, &params); err != nil {
		return types.BuildTxResult{}, fmt.Errorf("error loading account into wallet: %w", err)
	}
	return wallet.BuildTx(w.Wallet, params)
}

// SignTx signs a transaction
func (w *LocalWallet) SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network, &params); err != nil {
		return types.SignTxResult{}, fmt.Errorf("error signing tx: %w", err)
	}

	return wallet.SignTx(w.Wallet, params)
}

// SendTx submits a signed transaction to the Network
func (w *LocalWallet) SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error) {
	if err := w.loadAccountIntoWallet(ctx, params.Account, params.Network, &params); err != nil {
		return types.SendTxResult{}, fmt.Errorf("error loading account into wallet: %w", err)
	}

	return wallet.SendTx(w.Wallet, params)
}

// AddAccount adds an Account to the wallet
func (w *LocalWallet) AddAccount(acc account.Account) {
	w.accounts = append(w.accounts, acc)
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
