// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
)

func getDefaultSubnetEVMGenesis() subnet.SubnetParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("INITIAL_ALLOCATION_ADDRESS")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	return subnet.SubnetParams{
		SubnetEVM: &subnet.SubnetEVMParams{
			ChainID:     big.NewInt(123456),
			FeeConfig:   vm.StarterFeeConfig,
			Allocation:  allocation,
			Precompiles: params.Precompiles{},
		},
		Name: "TestSubnet",
	}
}

func DeploySubnet() {
	subnetParams := getDefaultSubnetEVMGenesis()
	// Create new Subnet EVM genesis
	newSubnet, _ := subnet.New(&subnetParams)

	network := avalanche.FujiNetwork()

	// Key that will be used for paying the transaction fees of CreateSubnetTx and CreateChainTx
	// NewKeychain will generate a new key pair in the provided path if no .pk file currently
	// exists in the provided path
	keychain, _ := keychain.NewKeychain(network, "KEY_PATH")

	// In this example, we are using the fee-paying key generated above also as control key
	// and subnet auth key

	// control keys are a list of keys that are permitted to make changes to a Subnet
	// such as creating a blockchain in the Subnet and adding validators to the Subnet
	controlKeys := keychain.Addresses().List()

	// subnet auth keys are a subset of control keys
	//
	// they are the keys that will be used to sign transactions that modify a Subnet
	// number of keys in subnetAuthKeys has to be more than or equal to threshold
	// all keys in subnetAuthKeys have to sign the transaction before the transaction
	// can be committed on chain
	subnetAuthKeys := keychain.Addresses().List()
	threshold := 1
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))

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

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	// since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, _ := newSubnet.Commit(*deployChainTx, wallet, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
