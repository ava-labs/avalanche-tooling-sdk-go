//go:build create_chain
// +build create_chain

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/blockchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/local"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func CreateChain(subnetID string) error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()
	network := network.FujiNetwork()

	localWallet, err := local.NewLocalWallet()
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	existingAccount, err := localWallet.ImportAccount("EXISTING_KEY_PATH")
	if err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}

	evmGenesisParams := blockchain.GetDefaultSubnetEVMGenesis("EVM_ADDRESS")
	evmGenesisBytes, _ := blockchain.CreateEvmGenesis(&evmGenesisParams)
	blockchainName := "TestBlockchain"
	vmID, err := blockchain.VMID(blockchainName)
	if err != nil {
		return fmt.Errorf("failed to get vmid: %w", err)
	}

	createChainParams := &pchainTxs.CreateChainTxParams{
		SubnetAuthKeys: []string{"P-fujixxxxx"},
		SubnetID:       subnetID,
		VMID:           vmID.String(),
		ChainName:      blockchainName,
		Genesis:        evmGenesisBytes,
	}
	buildTxParams := types.BuildTxParams{
		BaseParams: types.BaseParams{
			Account: *existingAccount,
			Network: network,
		},
		BuildTxInput: createChainParams,
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
	subnetID := "SUBNET_ID"
	if err := CreateChain(subnetID); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
