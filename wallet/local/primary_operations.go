// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/chains/pchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
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

	// Determine which wallet and account to use
	var walletToUse *primary.Wallet
	var accountToUse account.Account
	if len(params.AccountNames) > 0 {
		// Create temporary wallet for the specified account
		accountName := params.AccountNames[0]
		acc, exists := p.localWallet.accounts[accountName]
		if !exists {
			return types.BuildTxResult{}, fmt.Errorf("account %q not found", accountName)
		}
		tempWallet, err := getWalletFromAccount(ctx, acc, p.localWallet.defaultNetwork)
		if err != nil {
			return types.BuildTxResult{}, fmt.Errorf("error creating wallet for account: %w", err)
		}
		walletToUse = tempWallet
		accountToUse = acc
	} else {
		// Use the existing wallet (with active account)
		walletToUse = p.localWallet.wallet
		accountToUse = p.localWallet.accounts[p.localWallet.activeAccount]
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case pchain.ChainType:
		result, err := pchain.BuildTx(walletToUse, accountToUse, params)
		if err != nil {
			return types.BuildTxResult{}, err
		}
		return types.BuildTxResult{BuildTxOutput: result.BuildTxOutput}, nil
	default:
		return types.BuildTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

// SignTx signs a transaction
func (p *primaryOperations) SignTx(ctx context.Context, params types.SignTxParams) (types.SignTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SignTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Determine which wallet to use
	var walletToUse *primary.Wallet
	if len(params.AccountNames) > 0 {
		// Create temporary wallet for the specified account
		accountName := params.AccountNames[0]
		acc, exists := p.localWallet.accounts[accountName]
		if !exists {
			return types.SignTxResult{}, fmt.Errorf("account %q not found", accountName)
		}
		tempWallet, err := getWalletFromAccount(ctx, acc, p.localWallet.defaultNetwork)
		if err != nil {
			return types.SignTxResult{}, fmt.Errorf("error creating wallet for account: %w", err)
		}
		walletToUse = tempWallet
	} else {
		// Use the existing wallet (with active account)
		walletToUse = p.localWallet.wallet
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case pchain.ChainType:
		result, err := pchain.SignTx(walletToUse, params)
		if err != nil {
			return types.SignTxResult{}, err
		}
		return types.SignTxResult{SignTxOutput: result.SignTxOutput}, nil
	default:
		return types.SignTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

// SendTx submits a signed transaction to the Network
func (p *primaryOperations) SendTx(ctx context.Context, params types.SendTxParams) (types.SendTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SendTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Determine which wallet to use
	var walletToUse *primary.Wallet
	if len(params.AccountNames) > 0 {
		// Create temporary wallet for the specified account
		accountName := params.AccountNames[0]
		acc, exists := p.localWallet.accounts[accountName]
		if !exists {
			return types.SendTxResult{}, fmt.Errorf("account %q not found", accountName)
		}
		tempWallet, err := getWalletFromAccount(ctx, acc, p.localWallet.defaultNetwork)
		if err != nil {
			return types.SendTxResult{}, fmt.Errorf("error creating wallet for account: %w", err)
		}
		walletToUse = tempWallet
	} else {
		// Use the existing wallet (with active account)
		walletToUse = p.localWallet.wallet
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.SignTxResult.GetChainType(); chainType {
	case pchain.ChainType:
		result, err := pchain.SendTx(walletToUse, params)
		if err != nil {
			return types.SendTxResult{}, err
		}
		return types.SendTxResult{SendTxOutput: result.SendTxOutput}, nil
	default:
		return types.SendTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

// SubmitTx is a convenience method that combines BuildTx, SignTx, and SendTx
func (p *primaryOperations) SubmitTx(ctx context.Context, params types.SubmitTxParams) (types.SubmitTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return types.SubmitTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Determine which wallet and account to use (create once, reuse for all operations)
	var walletToUse *primary.Wallet
	var accountToUse account.Account
	if len(params.AccountNames) > 0 {
		// Create temporary wallet for the specified account
		accountName := params.AccountNames[0]
		acc, exists := p.localWallet.accounts[accountName]
		if !exists {
			return types.SubmitTxResult{}, fmt.Errorf("account %q not found", accountName)
		}
		tempWallet, err := getWalletFromAccount(ctx, acc, p.localWallet.defaultNetwork)
		if err != nil {
			return types.SubmitTxResult{}, fmt.Errorf("error creating wallet for account: %w", err)
		}
		walletToUse = tempWallet
		accountToUse = acc
	} else {
		// Use the existing wallet (with active account)
		walletToUse = p.localWallet.wallet
		accountToUse = p.localWallet.accounts[p.localWallet.activeAccount]
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case pchain.ChainType:
		// Step 1: Build the transaction
		buildParams := types.BuildTxParams(params)
		buildResult, err := pchain.BuildTx(walletToUse, accountToUse, buildParams)
		if err != nil {
			return types.SubmitTxResult{}, fmt.Errorf("failed to build tx: %w", err)
		}

		// Step 2: Sign the transaction
		signParams := types.SignTxParams{
			AccountNames:  params.AccountNames,
			BuildTxResult: &types.BuildTxResult{BuildTxOutput: buildResult.BuildTxOutput},
		}
		signResult, err := pchain.SignTx(walletToUse, signParams)
		if err != nil {
			return types.SubmitTxResult{}, fmt.Errorf("failed to sign tx: %w", err)
		}

		// Step 3: Send the transaction
		sendParams := types.SendTxParams{
			AccountNames: params.AccountNames,
			SignTxResult: &types.SignTxResult{SignTxOutput: signResult.SignTxOutput},
		}
		sendResult, err := pchain.SendTx(walletToUse, sendParams)
		if err != nil {
			return types.SubmitTxResult{}, fmt.Errorf("failed to send tx: %w", err)
		}

		return types.SubmitTxResult(sendResult), nil
	default:
		return types.SubmitTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}
