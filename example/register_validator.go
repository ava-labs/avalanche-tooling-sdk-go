// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func RegisterValidator(subnetID, chainID string) error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()
	network := network.FujiNetwork()

	localWallet, err := local.NewLocalWallet()
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	existingAccount, err := localWallet.ImportAccount("/Users/raymondsukanto/.avalanche-cli/key/newTestKey.pk")
	if err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}

	// Validator information
	nodeIDStr := "NodeID-7iXf6xZpXF8tFmYr1gGrEnPgK1sKqitAn"
	BLSPublicKey := "0xa569075008a3ce67ab6a1d7f9d3e9312c734cb99d144a169d98ac11d0177c9b3298a33d91e6b970745b802126edc5c5b"                                                                                                         // Replace with actual BLS public key
	BLSProofOfPossession := "0xaf1045946fd8668e492b7748e87bc20d5de232959b9fd7ea15f4f92c96f341ebeb1202c39b18148a5460d044655d25670a77e4079b255582153c3a372ab0b806e5f526bb1ff561f8f92eabe6e2b52adc491d1d36869818bb51be2c5b5329bcbb" // Replace with actual BLS proof
	ChangeOwnerAddr := "P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"
	ValidatorManagerAddress := "0x0FEEDC0DE0000000000000000000000000000000"
	Weight := 100         // Validator weight
	Balance := 1000000000 // Validator balance in nAVAX

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
		SubnetAuthKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		SubnetID:       subnetID,
		// ChainID is Blockchain ID of the L1 where the validator manager contract is deployed.
		ChainID: chainID,
		// Validators are the initial set of L1 validators after the conversion.
		Validators: bootstrapValidators,
		Address:    ValidatorManagerAddress,
	}
	buildTxParams := types.BuildTxParams{
		BaseParams: types.BaseParams{
			Account: *existingAccount,
			Network: network,
		},
		BuildTxInput: convertSubnetParams,
	}
	buildTxResult, err := localWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return fmt.Errorf("failed to BuildTx: %w", err)
	}

	signTxParams := types.SignTxParams{
		BaseParams: types.BaseParams{
			Account: *existingAccount,
			Network: network,
		},
		BuildTxResult: &buildTxResult,
	}
	signTxResult, err := localWallet.SignTx(ctx, signTxParams)
	if err != nil {
		return fmt.Errorf("failed to signTx: %w", err)
	}

	sendTxParams := types.SendTxParams{
		BaseParams: types.BaseParams{
			Account: *existingAccount,
			Network: network,
		},
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
	// Use a hardcoded subnet ID for this example
	// In a real scenario, you would get this from creating a subnet first
	subnetID := "2FmiyhpCzWpdiytmRJoazkxAfNowquCfSisHGB8rys13wZDRQz"
	chainID := "vzozrkRR95pvepywbV22oQWh7SqPb8FpegAKD71HfJSBwr3bM"
	if err := RegisterValidator(subnetID, chainID); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
