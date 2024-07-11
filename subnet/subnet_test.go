// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
)

func getDefaultSubnetEVMGenesis() SubnetParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("INITIAL_ALLOCATION_ADDRESS")] = core.GenesisAccount{
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

func TestSubnetDeploy(t *testing.T) {
	require := require.New(t)
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, err := New(&subnetParams)
	require.NoError(err)
	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "KEY_PATH")
	require.NoError(err)
	controlKeys := keychain.Addresses().List()
	subnetAuthKeys := keychain.Addresses().List()
	threshold := 1
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))
	wallet, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	require.NoError(err)
	deploySubnetTx, err := newSubnet.CreateSubnetTx(wallet)
	require.NoError(err)
	subnetID, err := newSubnet.Commit(*deploySubnetTx, wallet, true)
	require.NoError(err)
	fmt.Printf("subnetID %s \n", subnetID.String())
	time.Sleep(2 * time.Second)
	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(wallet)
	require.NoError(err)
	blockchainID, err := newSubnet.Commit(*deployChainTx, wallet, true)
	require.NoError(err)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func TestSubnetDeployMultiSig(t *testing.T) {
	require := require.New(t)
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := New(&subnetParams)
	network := avalanche.FujiNetwork()

	keychainA, err := keychain.NewKeychain(network, "KEY_PATH_A")
	require.NoError(err)
	keychainB, err := keychain.NewKeychain(network, "KEY_PATH_B")
	require.NoError(err)
	keychainC, err := keychain.NewKeychain(network, "KEY_PATH_C")
	require.NoError(err)

	controlKeys := []ids.ShortID{}
	controlKeys = append(controlKeys, keychainA.Addresses().List()[0])
	controlKeys = append(controlKeys, keychainB.Addresses().List()[0])
	controlKeys = append(controlKeys, keychainC.Addresses().List()[0])

	subnetAuthKeys := []ids.ShortID{}
	subnetAuthKeys = append(subnetAuthKeys, keychainA.Addresses().List()[0])
	subnetAuthKeys = append(subnetAuthKeys, keychainB.Addresses().List()[0])
	threshold := 2
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))

	walletA, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainA.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	require.NoError(err)

	deploySubnetTx, err := newSubnet.CreateSubnetTx(walletA)
	require.NoError(err)
	subnetID, err := newSubnet.Commit(*deploySubnetTx, walletA, true)
	require.NoError(err)
	fmt.Printf("subnetID %s \n", subnetID.String())

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	// first signature of CreateChainTx using keychain A
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	require.NoError(err)

	// include subnetID in PChainTxsToFetch when creating second wallet
	walletB, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainB.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	require.NoError(err)

	// second signature using keychain B
	err = walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx)
	require.NoError(err)

	// since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, err := newSubnet.Commit(*deployChainTx, walletA, true)
	require.NoError(err)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
