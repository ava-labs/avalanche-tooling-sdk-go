// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/blockchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
)

func CreateChain(subnetID string) error {
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
	evmGenesisParams := blockchain.GetDefaultSubnetEVMGenesis("initial_allocation_addr")
	evmGenesisBytes, _ := blockchain.CreateEvmGenesis(&evmGenesisParams)
	blockchainName := "TestBlockchain"
	vmID, err := blockchain.VmID(blockchainName)
	if err != nil {
		return fmt.Errorf("failed to get vmid: %w", err)
	}
	createSubnetParams := &txs.CreateChainTxParams{
		SubnetAuthKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		SubnetID:       subnetID,
		VMID:           vmID.String(),
		ChainName:      blockchainName,
		Genesis:        evmGenesisBytes,
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
	if sendTxResult.Tx != nil {
		fmt.Printf("sendTxResult %s \n", sendTxResult.Tx.ID())
	}
	return nil
}

func main() {
	// Run create chain example
	subnetID := "2b175hLJhG1m7HZ7aCLL4BTXFp2FEZsy5jfZ6wvFavr2Sx8g5n"
	if err := CreateChain(subnetID); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
