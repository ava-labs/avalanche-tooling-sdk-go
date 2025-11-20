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

func CreateSubnet() error {
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

	createSubnetParams := &pchainTxs.CreateSubnetTxParams{
		ControlKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		Threshold:   1,
	}
	buildTxParams := types.BuildTxParams{
		BuildTxInput: createSubnetParams,
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
	if err := CreateSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
