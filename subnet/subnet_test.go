// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
	"time"
)

func getDefaultSubnetEVMGenesis() SubnetParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("0x5a60e45535fbe925591cefb3e46f1052bbfcc67b")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	return SubnetParams{
		SubnetEVM: &SubnetEVMParams{
			ChainID:     big.NewInt(123456),
			FeeConfig:   vm.StarterFeeConfig,
			Allocation:  allocation,
			Precompiles: params.Precompiles{},
		},
		Name: "TestSubnet",
	}
}

func TestSubnetDeploy(_ *testing.T) {
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := New(&subnetParams)
	network := avalanche.FujiNetwork()
	keychain, _ := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk")
	controlKeys := keychain.Addresses().List()
	subnetAuthKeys := keychain.Addresses().List()
	threshold := 1
	newSubnet.SetDeployParams(controlKeys, subnetAuthKeys, uint32(threshold))
	wallet, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	deploySubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	subnetID, _ := newSubnet.Commit(*deploySubnetTx, wallet, true)
	fmt.Printf("subnetID %s \n", subnetID.String())
	time.Sleep(2 * time.Second)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	blockchainID, _ := newSubnet.Commit(*deployChainTx, wallet, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
