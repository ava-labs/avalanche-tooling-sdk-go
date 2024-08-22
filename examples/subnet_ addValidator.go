// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
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

func AddSubnetValidator() {
	ctx := context.Background()
	cp, err := node.GetDefaultCloudParams(ctx, node.AWSCloud)
	if err != nil {
		panic(err)
	}

	subnetParams := subnet.SubnetParams{
		GenesisFilePath: "/Users/raymondsukanto/.avalanche-cli/subnets/sdkSubnetNew/genesis.json",
		Name:            "sdkSubnetNew",
	}

	newSubnet, err := subnet.New(&subnetParams)
	if err != nil {
		panic(err)
	}

	node := node.Node{
		// NodeID is Avalanche Node ID of the node
		NodeID: "NodeID-Mb3AwcUpWysCWLP6mSpzzJVgYawJWzPHu",
		// IP address of the node
		IP: "18.144.79.215",
		// SSH configuration for the node
		SSHConfig: node.SSHConfig{
			User:           constants.RemoteHostUser,
			PrivateKeyPath: "/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem",
		},
		// Cloud is the cloud service that the node is on
		Cloud: node.AWSCloud,
		// CloudConfig is the cloud specific configuration for the node
		CloudConfig: *cp,
		// Role of the node can be 	Validator, API, AWMRelayer, Loadtest, or Monitor
		Roles: []node.SupportedRole{node.Validator},
	}

	subnetIDsToValidate := []string{newSubnet.SubnetID.String()}
	fmt.Printf("Reconfiguring node %s to track subnet %s\n", node.NodeID, subnetIDsToValidate)
	if err := node.SyncSubnets(subnetIDsToValidate); err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)

	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk", nil)
	if err != nil {
		panic(err)
	}

	subnetID, err := ids.FromString("2VsqBt64W9qayKttmGTiAmtsQVnp9e9U4gSHF1yuLKHuquck5j")

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

	//nodeID, err := ids.NodeIDFromString(node.NodeID)
	nodeID, err := ids.NodeIDFromString("NodeID-Mb3AwcUpWysCWLP6mSpzzJVgYawJWzPH")
	if err != nil {
		panic(err)
	}

	validator := subnet.SubnetValidatorParams{
		NodeID: nodeID,
		// Validate Subnet for 48 hours
		Duration: 48 * time.Hour,
		Weight:   20,
	}
	fmt.Printf("adding subnet validator")

	newSubnet.SetSubnetID(subnetID)
	subnetAuthKeys := keychain.Addresses().List()
	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)
	addValidatorTx, err := newSubnet.AddValidator(wallet, validator)
	if err != nil {
		panic(err)
	}
	txID, err := newSubnet.Commit(*addValidatorTx, wallet, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("obtained tx id %s", txID.String())
}
