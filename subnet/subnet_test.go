// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/teleporter"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"

	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func TestSubnetDeploy(_ *testing.T) {
	// Initialize a new Avalanche Object which will be used to set shared properties
	// like logging, metrics preferences, etc
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("0x5a60e45535fbe925591cefb3e46f1052bbfcc67b")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	teleporterInfo := &teleporter.Info{
		Version:                  "v1.0.0",
		FundedAddress:            "0x6e76EEf73Bcb65BCCd16d628Eb0B696552c53E4e",
		FundedBalance:            600000000000000000000,
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
	}
	newSubnet, _ := New(&subnetParams)
	ctx := context.Background()
	wallet, _ := wallet.New(
		ctx,
		&primary.WalletConfig{
			URI:              "",
			AVAXKeychain:     nil,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	// deploy Subnet returns multisig and error
	deploySubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	fmt.Printf("deploySubnetTx %s", deploySubnetTx)
}
