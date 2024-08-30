// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func DeploySubnetMultiSig() {
	subnetParams := getDefaultSubnetEVMGenesis()
	// Create new Subnet EVM genesis
	newSubnet, _ := subnet.New(&subnetParams)

	network := avalanche.FujiNetwork()

	// Create three keys that will be used as control keys of the subnet
	// NewKeychain will generate a new key pair in the provided path if no .pk file currently
	// exists in the provided path
	keychainA, _ := keychain.NewKeychain(network, "KEY_PATH_A", nil)
	keychainB, _ := keychain.NewKeychain(network, "KEY_PATH_B", nil)
	keychainC, _ := keychain.NewKeychain(network, "KEY_PATH_C", nil)

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
	// In this example, we are setting Key A and Key B as the subnet auth keys
	subnetAuthKeys = append(subnetAuthKeys, keychainA.Addresses().List()[0])
	subnetAuthKeys = append(subnetAuthKeys, keychainB.Addresses().List()[0])
	// at least two signatures are required to be able to send the CreateChain transaction on-chain
	// note that threshold does not apply to CreateSubnet transaction
	threshold := 2
	newSubnet.SetSubnetControlParams(controlKeys, uint32(threshold))

	// Key A will be used for paying the transaction fees of CreateSubnetTx and CreateChainTx
	walletA, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainA.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)

	deploySubnetTx, _ := newSubnet.CreateSubnetTx(walletA)
	subnetID, _ := newSubnet.Commit(*deploySubnetTx, walletA, true)
	fmt.Printf("subnetID %s \n", subnetID.String())

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	if err != nil {
		fmt.Errorf("error signing tx walletA: %w", err)
	}

	// we need to include subnetID in PChainTxsToFetch when creating second wallet
	// so that the wallet can fetch the CreateSubnet P-chain transaction to be able to
	// generate transactions.
	walletB, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainB.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)

	// sign with second wallet so that we have 2/2 threshold reached
	if err := walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx); err != nil {
		fmt.Errorf("error signing tx walletB: %w", err)
	}

	// since we have two signatures, we can now commit the transaction on chain
	blockchainID, _ := newSubnet.Commit(*deployChainTx, walletA, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
