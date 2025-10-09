//go:build create_chain
// +build create_chain

// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params/extras"

	"github.com/ava-labs/avalanche-tooling-sdk-go/blockchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
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

	existingAccount, err := localWallet.ImportAccount("/Users/raymondsukanto/.avalanche-cli/key/newTestKey.pk")
	if err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}
	evmGenesisParams := getDefaultSubnetEVMGenesis("initial_allocation_addr")
	evmGenesisBytes, _ := blockchain.CreateEvmGenesis(&evmGenesisParams)
	blockchainName := "TestBlockchain"
	vmID, err := blockchain.VmID(blockchainName)
	if err != nil {
		return fmt.Errorf("failed to get vmid: %w", err)
	}
	createChainParams := &pchainTxs.CreateChainTxParams{
		SubnetAuthKeys: []string{"P-fuji1377nx80rx3pzneup5qywgdgdsmzntql7trcqlg"},
		SubnetID:       subnetID,
		VMID:           vmID.String(),
		ChainName:      "TestBlockchain",
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

func getDefaultSubnetEVMGenesis(initialAllocationAddress string) blockchain.SubnetEVMParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress(initialAllocationAddress)] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	return blockchain.SubnetEVMParams{
		ChainID:     big.NewInt(123456),
		FeeConfig:   vm.StarterFeeConfig,
		Allocation:  allocation,
		Precompiles: extras.Precompiles{},
	}
}

func main() {
	// Use a hardcoded subnet ID for this example
	// In a real scenario, you would get this from creating a subnet first
	subnetID := "2b175hLJhG1m7HZ7aCLL4BTXFp2FEZsy5jfZ6wvFavr2Sx8g5n"
	if err := CreateChain(subnetID); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
