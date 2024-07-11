// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	"math/big"
	"testing"
	"time"

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

func TestSubnetDeployLedger(_ *testing.T) {
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := New(&subnetParams)
	network := avalanche.FujiNetwork()
	//fee := network.GenesisParams().CreateBlockchainTxFee + network.GenesisParams().CreateSubnetTxFee
	//ledgerInfo := keychain.LedgerParams{
	//	RequiredFunds: fee,
	//}
	ledgerInfo := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fuji1c8qchzh27navtly08kz5deenm5jvdqayyasuhh"}, //index 0
	}
	keychainA, err := keychain.NewKeychain(network, "", &ledgerInfo)
	fmt.Printf("error NewKeychain %s \n", err)
	addressesIDs, err := address.ParseToIDs([]string{"P-fuji1jjrlhfrx6e975y6e3fass72a09xp73zxd9gqf0"}) //index 7
	fmt.Printf("error ParseToIDs %s \n", err)
	controlKeys := addressesIDs
	fmt.Printf("control keys %s \n", controlKeys)
	subnetAuthKeys := addressesIDs
	fmt.Printf("subnetAuthKeys %s \n", subnetAuthKeys)
	threshold := 1
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))
	walletA, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainA.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	deploySubnetTx, err := newSubnet.CreateSubnetTx(walletA)
	fmt.Printf("error CreateSubnetTx %s \n", err)
	subnetID, err := newSubnet.Commit(*deploySubnetTx, walletA, true)
	fmt.Printf("error CreateSubnetTx Commit %s \n", err)
	fmt.Printf("subnetID %s \n", subnetID.String())
	time.Sleep(2 * time.Second)
	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	fmt.Printf("error CreateBlockchainTx %s \n", err)
	// include subnetID in PChainTxsToFetch when creating second wallet
	ledgerInfoB := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fuji1jjrlhfrx6e975y6e3fass72a09xp73zxd9gqf0"}, //index 7
	}
	err = keychainA.Ledger.LedgerDevice.Disconnect()
	fmt.Printf("error keychain A disconnect %s \n", err)
	keychainB, err := keychain.NewKeychain(network, "", &ledgerInfoB)
	fmt.Printf("error NewKeychainB %s \n", err)
	walletB, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainB.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	fmt.Printf("error new wallet B %s \n", err)

	// second signature
	if err = walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx); err != nil {
		fmt.Printf("error signing tx walletB: %s", err)
	}
	blockchainID, err := newSubnet.Commit(*deployChainTx, walletB, true)
	fmt.Printf("error commit %s \n", err)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
