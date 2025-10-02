// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func CreateSubnet() error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()
	network := network.FujiNetwork()

	localWallet, err := local.NewLocalWallet()
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	existingAccount, err := localWallet.ImportAccount(ctx, "/Users/raymondsukanto/.avalanche-cli/key/newTestKey.pk")
	if err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}

	createSubnetParams := &txs.CreateSubnetTxParams{
		ControlKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		Threshold:   1,
	}
	buildTxParams := wallet.BuildTxParams{
		BuildTxInput: createSubnetParams,
		Account:      *existingAccount,
		Network:      network,
	}
	buildTxResult, err := localWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return fmt.Errorf("failed to BuildTx: %w", err)
	}

	signTxParams := wallet.SignTxParams{
		BuildTxResult: &buildTxResult,
		Account:       *existingAccount,
		Network:       network,
	}
	signTxResult, err := localWallet.SignTx(ctx, signTxParams)
	if err != nil {
		return fmt.Errorf("failed to signTx: %w", err)
	}

	sendTxParams := wallet.SendTxParams{
		SignTxResult: &signTxResult,
		Account:      *existingAccount,
		Network:      network,
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
	if err := CreateSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
