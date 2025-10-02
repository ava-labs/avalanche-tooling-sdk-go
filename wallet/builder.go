// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/tx"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

// BuildTx constructs a transaction for the specified operation
func BuildTx(ctx context.Context, wallet *primary.Wallet, params BuildTxParams) (tx.BuildTxResult, error) {
	// Validate parameters first
	if err := params.Validate(); err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Route to appropriate chain handler based on chain type
	switch chainType := params.GetChainType(); chainType {
	case "P-Chain":
		return buildPChainTx(ctx, wallet, params.Account, params.BuildTxInput)
	case "C-Chain":
		return buildCChainTx(ctx, wallet, params)
	case "X-Chain":
		return buildXChainTx(ctx, wallet, params)
	default:
		return tx.BuildTxResult{}, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

func buildPChainTx(ctx context.Context, wallet *primary.Wallet, account account.Account, params BuildTxInput) (tx.BuildTxResult, error) {
	switch txType := params.GetTxType(); txType {
	case "CreateSubnetTx":
		createSubnetParams, ok := params.(*txs.CreateSubnetTxParams)
		if !ok {
			return tx.BuildTxResult{}, fmt.Errorf("invalid params type for ConvertSubnetToL1Tx, expected *txs.ConvertSubnetToL1TxParams")
		}
		return buildCreateSubnetTx(ctx, wallet, createSubnetParams)
	case "ConvertSubnetToL1Tx":
		convertParams, ok := params.(*txs.ConvertSubnetToL1TxParams)
		if !ok {
			return tx.BuildTxResult{}, fmt.Errorf("invalid params type for ConvertSubnetToL1Tx, expected *txs.ConvertSubnetToL1TxParams")
		}
		return buildConvertSubnetToL1Tx(ctx, wallet, account, convertParams)
	default:
		return tx.BuildTxResult{}, fmt.Errorf("unsupported P-Chain transaction type: %s", txType)
	}
}

// buildConvertSubnetToL1Tx provides a default implementation that can be used by any wallet
func buildConvertSubnetToL1Tx(ctx context.Context, wallet *primary.Wallet, account account.Account, params *txs.ConvertSubnetToL1TxParams) (tx.BuildTxResult, error) {
	options := getMultisigTxOptions(account, params.SubnetAuthKeys)
	unsignedTx, err := wallet.P().Builder().NewConvertSubnetToL1Tx(
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
	return *tx.NewPChainBuildTxResult(&builtTx), nil
}

// BuildCreateSubnetTx provides a default implementation that can be used by any wallet
func buildCreateSubnetTx(ctx context.Context, wallet *primary.Wallet, params *txs.CreateSubnetTxParams) (tx.BuildTxResult, error) {
	addrs, err := address.ParseToIDs(params.ControlKeys)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("failure parsing control keys: %w", err)
	}
	owners := &secp256k1fx.OutputOwners{
		Addrs:     addrs,
		Threshold: params.Threshold,
		Locktime:  0,
	}
	unsignedTx, err := wallet.P().Builder().NewCreateSubnetTx(
		owners,
	)
	if err != nil {
		return tx.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	return *tx.NewPChainBuildTxResult(&builtTx), nil
}

// getMultisigTxOptions is a helper function that can be shared
func getMultisigTxOptions(account account.Account, subnetAuthKeys []ids.ShortID) []common.Option {
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

// buildCChainTx builds C-Chain transactions
func buildCChainTx(ctx context.Context, wallet *primary.Wallet, params BuildTxInput) (tx.BuildTxResult, error) {
	// TODO: Implement C-Chain transaction building
	return tx.BuildTxResult{}, fmt.Errorf("C-Chain transactions not yet implemented")
}

// buildXChainTx builds X-Chain transactions
func buildXChainTx(ctx context.Context, wallet *primary.Wallet, params BuildTxInput) (tx.BuildTxResult, error) {
	// TODO: Implement X-Chain transaction building
	return tx.BuildTxResult{}, fmt.Errorf("X-Chain transactions not yet implemented")
}
