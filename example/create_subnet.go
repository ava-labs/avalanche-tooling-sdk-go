// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"os"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
)

func CreateSubnet() (ids.ID, error) {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()
	network := network.FujiNetwork()

	localWallet, err := local.NewLocalWallet()
	if err != nil {
		return ids.Empty, fmt.Errorf("failed to create wallet: %w", err)
	}

	existingAccount, err := localWallet.ImportAccount("/Users/raymondsukanto/.avalanche-cli/key/newTestKey.pk")
	if err != nil {
		return ids.Empty, fmt.Errorf("failed to ImportAccount: %w", err)
	}

	createSubnetParams := &pchainTxs.CreateSubnetTxParams{
		ControlKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		Threshold:   1,
	}
	buildTxParams := types.BuildTxParams{
		BaseParams: types.BaseParams{
			Account: *existingAccount,
			Network: network,
		},
		BuildTxInput: createSubnetParams,
	}
	buildTxResult, err := localWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return ids.Empty, fmt.Errorf("failed to BuildTx: %w", err)
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
		return ids.Empty, fmt.Errorf("failed to signTx: %w", err)
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
		return ids.Empty, fmt.Errorf("failed to sendTx: %w", err)
	}
	if tx := sendTxResult.GetTx(); tx != nil {
		if pChainTx, ok := tx.(*avagoTxs.Tx); ok {
			fmt.Printf("sendTxResult %s \n", pChainTx.ID())
			return pChainTx.ID(), nil
		}
	}
	return ids.Empty, fmt.Errorf("unable to get tx id")
}

func mainCreateSubnet() {
	if _, err := CreateSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
