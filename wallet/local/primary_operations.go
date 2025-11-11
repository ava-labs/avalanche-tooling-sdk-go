// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	txs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// primaryOperations implements the PrimaryOperations interface for LocalWallet
type primaryOperations struct {
	localWallet *LocalWallet
}

// Ensure primaryOperations implements PrimaryOperations interface
var _ wallet.PrimaryOperations = (*primaryOperations)(nil)

// BuildTx constructs a transaction for the specified operation
func (p *primaryOperations) BuildTx(ctx context.Context, params types.BuildTxParams) (types.BuildTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.BuildTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	walletToUse, err := p.getWalletToUse(ctx, params.AccountNames, &params)
	if err != nil {
		return types.BuildTxResult{}, err
	}

	var accountToUse account.Account
	if len(params.AccountNames) > 0 {
		accountToUse = p.localWallet.accounts[params.AccountNames[0]]
	} else {
		accountToUse = p.localWallet.accounts[p.localWallet.activeAccount]
	}

	return wallet.BuildTx(walletToUse, accountToUse, params)
}

// SignTx signs a transaction
func (p *primaryOperations) SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SignTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	walletToUse, err := p.getWalletToUse(ctx, params.AccountNames, &params)
	if err != nil {
		return types.SignTxResult{}, err
	}

	return wallet.SignTx(walletToUse, params)
}

// SendTx submits a signed transaction to the Network
func (p *primaryOperations) SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SendTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	walletToUse, err := p.getWalletToUse(ctx, params.AccountNames, &params)
	if err != nil {
		return types.SendTxResult{}, err
	}

	return wallet.SendTx(walletToUse, params)
}

// SubmitTx is a convenience method that combines BuildTx, SignTx, and SendTx
func (p *primaryOperations) SubmitTx(ctx context.Context, params types.SubmitTxParams) (types.SubmitTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Step 1: Build the transaction
	buildParams := types.BuildTxParams(params)
	buildResult, err := p.BuildTx(ctx, buildParams)
	if err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("failed to build tx: %w", err)
	}

	// Step 2: Sign the transaction
	signParams := types.SignTxParams{
		AccountNames:  params.AccountNames,
		BuildTxResult: &buildResult,
	}
	signResult, err := p.SignTx(ctx, signParams)
	if err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("failed to sign tx: %w", err)
	}

	// Step 3: Send the transaction
	sendParams := types.SendTxParams{
		AccountNames: params.AccountNames,
		SignTxResult: &signResult,
	}
	sendResult, err := p.SendTx(ctx, sendParams)
	if err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("failed to send tx: %w", err)
	}

	return types.SubmitTxResult(sendResult), nil
}

// hasSeenSubnetID checks if a subnet ID has been seen before
func (p *primaryOperations) hasSeenSubnetID(subnetID ids.ID) bool {
	for _, id := range p.localWallet.seenSubnetIDs {
		if id == subnetID {
			return true
		}
	}
	return false
}

// getWalletToUse determines which wallet to use based on AccountNames parameter
func (p *primaryOperations) getWalletToUse(ctx context.Context, accountNames []string, params interface{}) (*primary.Wallet, error) {
	if len(accountNames) > 0 {
		// Create temporary wallet for the specified account
		accountName := accountNames[0]
		acc, exists := p.localWallet.accounts[accountName]
		if !exists {
			return nil, fmt.Errorf("account %q not found", accountName)
		}
		config := buildWalletConfig(params)
		tempWallet, err := getWalletFromAccount(ctx, acc, p.localWallet.defaultNetwork, &config)
		if err != nil {
			return nil, fmt.Errorf("error creating wallet for account: %w", err)
		}
		return tempWallet, nil
	}
	// Use the existing wallet (with active account)
	if p.localWallet.activeAccount == "" {
		return nil, fmt.Errorf("no account specified and no active account set")
	}

	// Check if we need to refresh wallet with new subnet IDs
	config := buildWalletConfig(params)
	if len(config.SubnetIDs) > 0 {
		subnetID := config.SubnetIDs[0]
		if subnetID != ids.Empty && !p.hasSeenSubnetID(subnetID) {
			// New subnet ID detected, rebuild wallet with all seen subnet IDs
			p.localWallet.seenSubnetIDs = append(p.localWallet.seenSubnetIDs, subnetID)
			acc := p.localWallet.accounts[p.localWallet.activeAccount]
			newConfig := primary.WalletConfig{SubnetIDs: p.localWallet.seenSubnetIDs}
			newWallet, err := getWalletFromAccount(ctx, acc, p.localWallet.defaultNetwork, &newConfig)
			if err != nil {
				return nil, fmt.Errorf("error refreshing wallet with new subnet ID: %w", err)
			}
			p.localWallet.wallet = newWallet
		}
	}

	return p.localWallet.wallet, nil
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
