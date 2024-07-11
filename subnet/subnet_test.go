// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/utils/formatting/address"
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

func TestSubnetDeploy(_ *testing.T) {
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := New(&subnetParams)
	network := avalanche.FujiNetwork()
	keychain, _ := keychain.NewKeychain(network, "KEY_PATH", nil)
	controlKeys := keychain.Addresses().List()
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
	time.Sleep(2 * time.Second)
	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	blockchainID, _ := newSubnet.Commit(*deployChainTx, wallet, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func TestSubnetDeployLedger(t *testing.T) {
	require := require.New(t)
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, err := New(&subnetParams)
	require.NoError(err)
	network := avalanche.FujiNetwork()

	ledgerInfo := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fujixxxxxxxxx"},
	}
	keychainA, err := keychain.NewKeychain(network, "", &ledgerInfo)
	require.NoError(err)

	addressesIDs, err := address.ParseToIDs([]string{"P-fujiyyyyyyyy"})
	require.NoError(err)
	controlKeys := addressesIDs
	subnetAuthKeys := addressesIDs
	threshold := 1

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

	time.Sleep(2 * time.Second)

	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	require.NoError(err)

	ledgerInfoB := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fujiyyyyyyyy"},
	}
	err = keychainA.Ledger.LedgerDevice.Disconnect()
	require.NoError(err)

	keychainB, err := keychain.NewKeychain(network, "", &ledgerInfoB)
	require.NoError(err)

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

	// second signature
	err = walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx)
	require.NoError(err)

	blockchainID, err := newSubnet.Commit(*deployChainTx, walletB, true)
	require.NoError(err)

	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
