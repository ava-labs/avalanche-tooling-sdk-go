// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
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

func TestSubnetDeploy(_ *testing.T) {
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := New(&subnetParams)
	network := avalanche.FujiNetwork()
	keychain, _ := keychain.NewKeychain("KEY_PATH")
	controlKeys := keychain.Addresses().List()
	subnetAuthKeys := keychain.Addresses().List()
	threshold := 1
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))
	wallet, _ := wallet.New(
		context.Background(),
		network,
		keychain,
	)
	deploySubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	subnetID, _ := newSubnet.Commit(deploySubnetTx, wallet, true)
	fmt.Printf("subnetID %s \n", subnetID.String())
	time.Sleep(2 * time.Second)
	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	blockchainID, _ := newSubnet.Commit(deployChainTx, wallet, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func TestSubnetDeployMultiSig(t *testing.T) {
	require := require.New(t)

	subnetParams := getDefaultSubnetEVMGenesis()
	// Create new Subnet EVM genesis
	newSubnet, _ := New(&subnetParams)

	network := avalanche.FujiNetwork()

	// Key that will be used for paying the transaction fees of CreateSubnetTx and CreateChainTx
	// NewKeychain will generate a new key pair in the provided path if no .pk file currently
	// exists in the provided path
	keyA, err := key.LoadEwoq()
	require.NoError(err)
	keyB, err := key.NewSoft()
	require.NoError(err)
	keyC, err := key.NewSoft()
	require.NoError(err)
	keychainA := keychain.KeychainFromKey(keyA)
	keychainB := keychain.KeychainFromKey(keyB)
	keychainC := keychain.KeychainFromKey(keyC)

	// In this example, we are using the fee-paying key generated above also as control key
	// and subnet auth key

	// control keys are a list of keys that are permitted to make changes to a Subnet
	// such as creating a blockchain in the Subnet and adding validators to the Subnet
	controlKeys := []ids.ShortID{}
	controlKeys = append(controlKeys, keychainA.Addresses().List()[0])
	controlKeys = append(controlKeys, keychainB.Addresses().List()[0])
	controlKeys = append(controlKeys, keychainC.Addresses().List()[0])

	// subnet auth keys are a subset of control keys
	//
	// they are the keys that will be used to sign transactions that modify a Subnet
	// number of keys in subnetAuthKeys has to be more than or equal to threshold
	// all keys in subnetAuthKeys have to sign the transaction before the transaction
	// can be committed on chain
	subnetAuthKeys := []ids.ShortID{}
	subnetAuthKeys = append(subnetAuthKeys, keychainA.Addresses().List()[0])
	subnetAuthKeys = append(subnetAuthKeys, keychainB.Addresses().List()[0])
	threshold := 2
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))

	ctx, cancel := utils.GetAPIContext()
	defer cancel()
	walletA, err := wallet.New(
		ctx,
		network,
		keychainA,
	)
	require.NoError(err)

	deploySubnetTx, err := newSubnet.CreateSubnetTx(walletA)
	require.NoError(err)

	subnetID, err := newSubnet.Commit(deploySubnetTx, walletA, true)
	require.NoError(err)

	fmt.Printf("subnetID %s \n", subnetID.String())

	ctx, cancel = utils.GetAPIContext()
	defer cancel()
	walletB, err := wallet.New(
		ctx,
		network,
		keychainB,
	)
	require.NoError(err)

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	require.NoError(err)
	_, remainingSigners, err := deployChainTx.GetRemainingAuthSigners()
	require.NoError(err)
	remainingSignersPFormat, err := utils.P(network.HRP(), remainingSigners)
	require.NoError(err)
	fmt.Printf("remainingSigners %s\n", remainingSignersPFormat)

	fmt.Printf("signing with wallet B \n")
	ctx, cancel = utils.GetAPIContext()
	defer cancel()
	err = walletB.Sign(ctx, deployChainTx)
	require.NoError(err)

	_, remainingSigners, err = deployChainTx.GetRemainingAuthSigners()
	require.NoError(err)
	remainingSignersPFormat, err = utils.P(network.HRP(), remainingSigners)
	require.NoError(err)
	fmt.Printf("remainingSigners %s\n", remainingSignersPFormat)

	// since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, err := newSubnet.Commit(deployChainTx, walletA, true)
	require.NoError(err)
	fmt.Printf("blockchainID %s\n", blockchainID.String())
}
