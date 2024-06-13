// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/teleporter"
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

func TestSubnetDeploy(_ *testing.T) {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("0x5a60e45535fbe925591cefb3e46f1052bbfcc67b")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	teleporterInfo := &teleporter.Info{
		Version:                  "v1.0.0",
		FundedAddress:            "0x6e76EEf73Bcb65BCCd16d628Eb0B696552c53E4e",
		FundedBalance:            big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(600)),
		MessengerDeployerAddress: "0x618FEdD9A45a8C456812ecAAE70C671c6249DfaC",
		RelayerAddress:           "0x2A20d1623ce3e90Ec5854c84E508B8af065C059d",
	}
	subnetParams := SubnetParams{
		SubnetEVM: &SubnetEVMParams{
			EnableWarp:     true,
			ChainID:        big.NewInt(123456),
			FeeConfig:      vm.StarterFeeConfig,
			Allocation:     allocation,
			Precompiles:    params.Precompiles{},
			TeleporterInfo: teleporterInfo,
		},
		Name: "TestSubnet",
	}
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
	subnetID, _ := wallet.Commit(*deploySubnetTx, true)
	fmt.Printf("subnetID %s \n", subnetID.String())
	newSubnet.SubnetID = subnetID
	time.Sleep(2 * time.Second)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	blockchainID, _ := wallet.Commit(*deployChainTx, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
