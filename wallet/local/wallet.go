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
	primaryWallet     *primary.Wallet            // Avalanchego primary wallet for P/X/C operations
	accounts          map[string]account.Account // Named accounts map
	activeAccountName string                     // Currently active account name
	activeNetwork     network.Network            // Active network for operations
	seenSubnetIDs     []ids.ID                   // Subnet IDs seen for active account
}

// Ensure LocalWallet implements Wallet interface
var _ wallet.Wallet = (*LocalWallet)(nil)

// NewLocalWallet creates a new local wallet with the specified network
func NewLocalWallet(network network.Network) (*LocalWallet, error) {
	return &LocalWallet{
		primaryWallet:     nil,
		accounts:          make(map[string]account.Account),
		activeAccountName: "",
		activeNetwork:     network,
		seenSubnetIDs:     []ids.ID{},
	}, nil
}

// SetNetwork sets the active network for wallet operations
func (w *LocalWallet) SetNetwork(network network.Network) {
	w.activeNetwork = network
}

// Network returns the active network for wallet operations
func (w *LocalWallet) Network() network.Network {
	return w.activeNetwork
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
			// For ConvertSubnetToL1Tx, the subnet ID field is Subnet
			return unsignedTx.Subnet, nil
		}
	}
	return ids.Empty, fmt.Errorf("no subnet ID found in transaction")
}

// getWalletToUse determines which wallet to use based on AccountNames parameter
func (w *LocalWallet) getWalletToUse(ctx context.Context, accountNames []string, params interface{}) (*primary.Wallet, error) {
	if len(accountNames) > 0 {
		// Create temporary wallet for the specified account
		// TODO: Support multiple accounts for multisig transactions
		accountName := accountNames[0]
		acc, exists := w.accounts[accountName]
		if !exists {
			return nil, fmt.Errorf("account %q not found", accountName)
		}
		config := buildWalletConfig(params)
		tempWallet, err := createPrimaryWallet(ctx, acc, w.activeNetwork, &config)
		if err != nil {
			return nil, fmt.Errorf("error creating wallet for account: %w", err)
		}
		return tempWallet, nil
	}
	// Use the existing wallet (with active account)
	if w.activeAccountName == "" {
		return nil, fmt.Errorf("no account specified and no active account set")
	}

	// Check if we need to refresh wallet with new subnet IDs
	config := buildWalletConfig(params)
	if len(config.SubnetIDs) > 0 {
		subnetID := config.SubnetIDs[0]
		if subnetID != ids.Empty && !w.hasSeenSubnetID(subnetID) {
			// New subnet ID detected, rebuild wallet with all seen subnet IDs
			w.seenSubnetIDs = append(w.seenSubnetIDs, subnetID)
			acc := w.accounts[w.activeAccountName]
			newConfig := primary.WalletConfig{SubnetIDs: w.seenSubnetIDs}
			newWallet, err := createPrimaryWallet(ctx, acc, w.activeNetwork, &newConfig)
			if err != nil {
				return nil, fmt.Errorf("error refreshing wallet with new subnet ID: %w", err)
			}
			w.primaryWallet = newWallet
		}
	}

	return w.primaryWallet, nil
}

// hasSeenSubnetID checks if a subnet ID has been seen before
func (w *LocalWallet) hasSeenSubnetID(subnetID ids.ID) bool {
	for _, id := range w.seenSubnetIDs {
		if id == subnetID {
			return true
		}
	}
	return false
}

// BuildTx constructs a transaction for the specified operation
func (w *LocalWallet) BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error) {
	if err := params.Validate(); err != nil {
		return types.BuildTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	walletToUse, err := w.getWalletToUse(ctx, params.AccountNames, &params)
	if err != nil {
		return types.BuildTxResult{}, err
	}

	var accountToUse account.Account
	if len(params.AccountNames) > 0 {
		// TODO: Support multiple accounts for multisig transactions
		accountToUse = w.accounts[params.AccountNames[0]]
	} else {
		accountToUse = w.accounts[w.activeAccountName]
	}

	return wallet.BuildTx(walletToUse, accountToUse, params)
}

// SignTx signs a transaction
func (w *LocalWallet) SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error) {
	if err := params.Validate(); err != nil {
		return types.SignTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	walletToUse, err := w.getWalletToUse(ctx, params.AccountNames, &params)
	if err != nil {
		return types.SignTxResult{}, err
	}

	return wallet.SignTx(walletToUse, params)
}

// SendTx submits a signed transaction to the Network
func (w *LocalWallet) SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error) {
	if err := params.Validate(); err != nil {
		return types.SendTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	walletToUse, err := w.getWalletToUse(ctx, params.AccountNames, &params)
	if err != nil {
		return types.SendTxResult{}, err
	}

	return wallet.SendTx(walletToUse, params)
}

// SubmitTx builds, signs, and sends a transaction in one call
func (w *LocalWallet) SubmitTx(ctx context.Context, params types.SubmitTxParams) (types.SubmitTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Step 1: Build the transaction
	buildParams := types.BuildTxParams(params)
	buildResult, err := w.BuildTx(ctx, buildParams)
	if err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("failed to build tx: %w", err)
	}

	// Step 2: Sign the transaction
	signParams := types.SignTxParams{
		AccountNames:  params.AccountNames,
		BuildTxResult: &buildResult,
	}
	signResult, err := w.SignTx(ctx, signParams)
	if err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("failed to sign tx: %w", err)
	}

	// Step 3: Send the transaction
	sendParams := types.SendTxParams{
		AccountNames: params.AccountNames,
		SignTxResult: &signResult,
	}
	sendResult, err := w.SendTx(ctx, sendParams)
	if err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("failed to send tx: %w", err)
	}

	return types.SubmitTxResult(sendResult), nil
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
		issueTxErr = w.primaryWallet.P().IssueTx(tx, options...)
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
