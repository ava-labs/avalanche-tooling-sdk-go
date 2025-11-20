//go:build convert_subnet
// +build convert_subnet

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func ConvertSubnet(subnetID, chainID string) error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()
	network := network.FujiNetwork()

	localWallet, err := local.NewLocalWallet(network)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	acc, err := account.ImportFromPrivateKey("my-account", "EXISTING_PRIVATE_KEY_HEX")
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	if err = localWallet.ImportAccount(acc); err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}

	// Validator information
	nodeIDStr := "NodeID-xxxxx"
	BLSPublicKey := "0x....,."        // Replace with actual BLS public key
	BLSProofOfPossession := "0x....." // Replace with actual BLS proof of possession
	ChangeOwnerAddr := "P-fujixxxx"
	ValidatorManagerAddress := "0x0FEEDC0DE0000000000000000000000000000000" // default validator manager address
	Weight := 100                                                           // Validator weight
	Balance := 1000000000                                                   // Validator balance in nAVAX

	bootstrapValidators := []*pchainTxs.ConvertSubnetToL1Validator{}
	bootstrapValidator := &pchainTxs.ConvertSubnetToL1Validator{
		NodeID:                nodeIDStr,
		Weight:                uint64(Weight),
		Balance:               uint64(Balance),
		BLSPublicKey:          BLSPublicKey,
		BLSProofOfPossession:  BLSProofOfPossession,
		RemainingBalanceOwner: ChangeOwnerAddr,
	}
	bootstrapValidators = append(bootstrapValidators, bootstrapValidator)

	convertSubnetParams := &pchainTxs.ConvertSubnetToL1TxParams{
		SubnetAuthKeys: []string{"P-fujixxxxxx"},
		SubnetID:       subnetID,
		// ChainID is Blockchain ID of the L1 where the validator manager contract is deployed.
		ChainID: chainID,
		// Validators are the initial set of L1 validators after the conversion.
		Validators: bootstrapValidators,
		Address:    ValidatorManagerAddress,
	}
	buildTxParams := types.BuildTxParams{
		BuildTxInput: convertSubnetParams,
	}
	buildTxResult, err := localWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return fmt.Errorf("failed to BuildTx: %w", err)
	}

	signTxParams := types.SignTxParams{
		BuildTxResult: &buildTxResult,
	}
	signTxResult, err := localWallet.SignTx(ctx, signTxParams)
	if err != nil {
		return fmt.Errorf("failed to signTx: %w", err)
	}

	sendTxParams := types.SendTxParams{
		SignTxResult: &signTxResult,
	}
	sendTxResult, err := localWallet.SendTx(ctx, sendTxParams)
	if err != nil {
		return fmt.Errorf("failed to sendTx: %w", err)
	}
	if tx := sendTxResult.GetTx(); tx != nil {
		if pChainTx, ok := tx.(*avagoTxs.Tx); ok {
			fmt.Printf("sendTxResult %s \n", pChainTx.ID())
		} else {
			fmt.Printf("sendTxResult %s transaction \n", sendTxResult.GetChainType())
		}
	}
	return nil
}

func main() {
	// Use a hardcoded subnet ID & chain ID for this example
	// In a real scenario, you would get this from creating a subnet & creating a blockchain first
	subnetID := "SUBNET_ID"
	chainID := "CHAIN_ID"
	if err := ConvertSubnet(subnetID, chainID); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
