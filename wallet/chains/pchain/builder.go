// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package pchain

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/message"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// BuildTx builds P-Chain transactions
func BuildTx(wallet *primary.Wallet, account account.Account, params types.BuildTxParams) (types.BuildTxResult, error) {
	switch txType := params.BuildTxInput.(type) {
	case *pchainTxs.CreateSubnetTxParams:
		return buildCreateSubnetTx(wallet, txType)
	case *pchainTxs.CreateChainTxParams:
		return buildCreateChainTx(wallet, account, txType)
	case *pchainTxs.ConvertSubnetToL1TxParams:
		return buildConvertSubnetToL1Tx(wallet, account, txType)
	default:
		return types.BuildTxResult{}, fmt.Errorf("unsupported P-Chain transaction type: %T", params.BuildTxInput)
	}
}

// buildCreateSubnetTx provides a default implementation that can be used by any wallet
func buildCreateSubnetTx(wallet *primary.Wallet, params *pchainTxs.CreateSubnetTxParams) (types.BuildTxResult, error) {
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

// buildCreateChainTx provides a default implementation that can be used by any wallet
func buildCreateChainTx(wallet *primary.Wallet, account account.Account, params *pchainTxs.CreateChainTxParams) (types.BuildTxResult, error) {
	subnetAuthKeys, err := convertSubnetAuthKeys(params.SubnetAuthKeys)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to convert subnet auth keys: %w", err)
	}
	options := getMultisigTxOptions(account, subnetAuthKeys)
	fxIDs := make([]ids.ID, 0)
	subnetID, err := ids.FromString(params.SubnetID)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to parse subnet ID: %w", err)
	}
	vmID, err := ids.FromString(params.VMID)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to parse VM ID: %w", err)
	}
	unsignedTx, err := wallet.P().Builder().NewCreateChainTx(
		subnetID,
		params.Genesis,
		vmID,
		fxIDs,
		params.ChainName,
		options...,
	)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("error building tx: %w", err)
	}
	builtTx := avagoTxs.Tx{Unsigned: unsignedTx}
	pChainResult := types.NewPChainBuildTxResult(&builtTx)
	return types.BuildTxResult{BuildTxOutput: pChainResult}, nil
}

// buildConvertSubnetToL1Tx provides a default implementation that can be used by any wallet
func buildConvertSubnetToL1Tx(wallet *primary.Wallet, account account.Account, params *pchainTxs.ConvertSubnetToL1TxParams) (types.BuildTxResult, error) {
	subnetAuthKeys, err := convertSubnetAuthKeys(params.SubnetAuthKeys)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to convert subnet auth keys: %w", err)
	}
	subnetID, err := ids.FromString(params.SubnetID)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to parse subnet ID: %w", err)
	}
	chainID, err := ids.FromString(params.ChainID)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to parse chain ID: %w", err)
	}
	addressBytes := []byte(params.Address)
	avagoValidators, err := convertSubnetValidatorsToAvagoValidators(params.Validators)
	if err != nil {
		return types.BuildTxResult{}, fmt.Errorf("failed to convert address to bytes: %w", err)
	}
	options := getMultisigTxOptions(account, subnetAuthKeys)
	unsignedTx, err := wallet.P().Builder().NewConvertSubnetToL1Tx(
		subnetID,
		chainID,
		addressBytes,
		avagoValidators,
		options...,
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

// convertSubnetAuthKeys converts a slice of string addresses to a slice of ShortIDs
func convertSubnetAuthKeys(subnetAuthKeys []string) ([]ids.ShortID, error) {
	subnetAuthKeyIDs, err := address.ParseToIDs(subnetAuthKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to convert subnet auth key %s to ShortID: %w", subnetAuthKeys, err)
	}
	return subnetAuthKeyIDs, nil
}

func convertSubnetValidatorsToAvagoValidators(validators []*pchainTxs.ConvertSubnetToL1Validator) ([]*avagoTxs.ConvertSubnetToL1Validator, error) {
	bootstrapValidators := []*avagoTxs.ConvertSubnetToL1Validator{}
	for _, validator := range validators {
		nodeID, err := ids.NodeIDFromString(validator.NodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node ID: %w", err)
		}

		blsInfo, err := convertToBLSProofOfPossession(validator.BLSPublicKey, validator.BLSProofOfPossession)
		if err != nil {
			return nil, fmt.Errorf("failure parsing BLS info: %w", err)
		}

		addrs, err := address.ParseToIDs([]string{validator.RemainingBalanceOwner})
		if err != nil {
			return nil, fmt.Errorf("failure parsing change owner address: %w", err)
		}
		bootstrapValidator := &avagoTxs.ConvertSubnetToL1Validator{
			NodeID:  nodeID[:],
			Weight:  validator.Weight,
			Balance: validator.Balance,
			Signer:  blsInfo,
			RemainingBalanceOwner: message.PChainOwner{
				Threshold: 1,
				Addresses: addrs,
			},
		}
		bootstrapValidators = append(bootstrapValidators, bootstrapValidator)
	}

	return bootstrapValidators, nil
}

func convertToBLSProofOfPossession(publicKey, proofOfPossesion string) (signer.ProofOfPossession, error) {
	type jsonProofOfPossession struct {
		PublicKey         string
		ProofOfPossession string
	}
	jsonPop := jsonProofOfPossession{
		PublicKey:         publicKey,
		ProofOfPossession: proofOfPossesion,
	}
	popBytes, err := json.Marshal(jsonPop)
	if err != nil {
		return signer.ProofOfPossession{}, err
	}
	pop := &signer.ProofOfPossession{}
	err = pop.UnmarshalJSON(popBytes)
	if err != nil {
		return signer.ProofOfPossession{}, err
	}
	return *pop, nil
}
