// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func TestValidateSubnet(_ *testing.T) {
	subnetParams := SubnetParams{
		GenesisFilePath: "/Users/raymondsukanto/.avalanche-cli/subnets/sdkSubnetNew/genesis.json",
		Name:            "sdkSubnetNew",
	}

	newSubnet, err := New(&subnetParams)
	if err != nil {
		panic(err)
	}

	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk", nil)
	if err != nil {
		panic(err)
	}

	subnetID, err := ids.FromString("2VsqBt64W9qayKttmGTiAmtsQVnp9e9U4gSHF1yuLKHuquck5j")
	if err != nil {
		panic(err)
	}

	wallet, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	if err != nil {
		panic(err)
	}

	nodeID, err := ids.NodeIDFromString("NodeID-Mb3AwcUpWysCWLP6mSpzzJVgYawJWzPH")
	if err != nil {
		panic(err)
	}

	validator := validator.SubnetValidatorParams{
		NodeID: nodeID,
		// Validate Subnet for 48 hours
		Duration: 48 * time.Hour,
		Weight:   20,
	}
	fmt.Printf("adding subnet validator")

	newSubnet.SetSubnetID(subnetID)

	// We need to set Subnet Auth Keys for this transaction since Subnet AddValidator is
	// a Subnet-changing transaction
	//
	// In this example, the example Subnet was created with only 1 key as control key with a threshold of 1
	// and the control key is the key contained in the keychain object, so we are going to use the
	// key contained in the keychain object as the Subnet Auth Key for Subnet AddValidator tx
	subnetAuthKeys := keychain.Addresses().List()
	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)

	// In this example, we are assuming that the specified node has already tracked the Subnet
	// which can be done in Avalanche Tooling SDK through node.Sync function
	// For an example, head to examples/subnet_addValidator.go
	addValidatorTx, err := newSubnet.AddValidator(wallet, validator)
	if err != nil {
		panic(err)
	}

	// Since it has the required signatures, we will now commit the transaction on chain
	txID, err := newSubnet.Commit(*addValidatorTx, wallet, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("obtained tx id %s", txID.String())
}
