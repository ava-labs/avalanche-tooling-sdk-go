// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
)

func CreateSubnet() error {

	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()
	network := network.FujiNetwork()

	wallet, err := NewLocalWallet()
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	existingAccount, err := wallet.ImportAccount(ctx, "/Users/raymondsukanto/.avalanche-cli/key/newTestKey.pk")
	if err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}

	createSubnetParams := &txs.CreateSubnetTxParams{
		ControlKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		Threshold:   1,
	}
	buildTxParams := BuildTxParams{
		BuildTxInput: createSubnetParams,
		account:      *existingAccount,
		network:      network,
	}
	buildTxResult, err := wallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return fmt.Errorf("failed to BuildTx: %w", err)
	}

	signTxParams := SignTxParams{
		BuildTxResult: &buildTxResult,
		account:       *existingAccount,
		network:       network,
	}
	signTxResult, err := wallet.SignTx(ctx, signTxParams)
	if err != nil {
		return fmt.Errorf("failed to signTx: %w", err)
	}

	sendTxParams := SendTxParams{
		SignTxResult: &signTxResult,
		account:      *existingAccount,
		network:      network,
	}
	sendTxResult, err := wallet.SendTx(ctx, sendTxParams)
	if err != nil {
		return fmt.Errorf("failed to signTx: %w", err)
	}
	fmt.Printf("sendTxResult %s \n", sendTxResult.Tx.ID())
	return nil
}

func main() {
	if err := CreateSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
