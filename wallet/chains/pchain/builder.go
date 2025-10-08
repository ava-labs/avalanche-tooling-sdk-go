// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package pchain

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// BuildTx builds P-Chain transactions
func BuildTx(ctx context.Context, wallet *primary.Wallet, account account.Account, params types.BuildTxParams) (types.BuildTxResult, error) {
	switch txType := params.BuildTxInput.(type) {
	case *pchainTxs.CreateSubnetTxParams:
		return buildCreateSubnetTx(ctx, wallet, txType)
	case *pchainTxs.ConvertSubnetToL1TxParams:
		return buildConvertSubnetToL1Tx(ctx, wallet, account, txType)
	default:
		return types.BuildTxResult{}, fmt.Errorf("unsupported P-Chain transaction type: %T", params.BuildTxInput)
	}
}

// buildConvertSubnetToL1Tx provides a default implementation that can be used by any wallet
func buildConvertSubnetToL1Tx(ctx context.Context, wallet *primary.Wallet, account account.Account, params *pchainTxs.ConvertSubnetToL1TxParams) (types.BuildTxResult, error) {
	options := getMultisigTxOptions(account, params.SubnetAuthKeys)
	unsignedTx, err := wallet.P().Builder().NewConvertSubnetToL1Tx(
		params.SubnetID,
		params.ChainID,
		params.Address,
		params.Validators,
		options...,
	)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	pChainResult := types.NewPChainBuildTxResult(&builtTx)
	return types.BuildTxResult{BuildTxOutput: pChainResult}, nil
}

// buildCreateSubnetTx provides a default implementation that can be used by any wallet
func buildCreateSubnetTx(ctx context.Context, wallet *primary.Wallet, params *pchainTxs.CreateSubnetTxParams) (types.BuildTxResult, error) {
	addrs, err := address.ParseToIDs(params.ControlKeys)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failure parsing control keys: %w", err)
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
		return types.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	pChainResult := types.NewPChainBuildTxResult(&builtTx)
	return types.BuildTxResult{BuildTxOutput: pChainResult}, nil
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
