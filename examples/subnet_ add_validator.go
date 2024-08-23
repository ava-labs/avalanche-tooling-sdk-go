// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"time"
)

func AddSubnetValidator() {
	// We are using existing Subnet that we have already deployed on Fuji
	subnetParams := subnet.SubnetParams{
		GenesisFilePath: "GENESIS_FILE_PATH",
		Name:            "SUBNET_NAME",
	}

	newSubnet, err := subnet.New(&subnetParams)
	if err != nil {
		panic(err)
	}

	subnetID, err := ids.FromString("SUBNET_ID")
	if err != nil {
		panic(err)
	}

	// Genesis doesn't contain the deployed Subnet's SubnetID, we need to first set the Subnet ID
	newSubnet.SetSubnetID(subnetID)

	// We are using existing host
	node := node.Node{
		// NodeID is Avalanche Node ID of the node
		NodeID: "NODE_ID",
		// IP address of the node
		IP: "NODE_IP_ADDRESS",
		// SSH configuration for the node
		SSHConfig: node.SSHConfig{
			User:           constants.RemoteHostUser,
			PrivateKeyPath: "NODE_KEYPAIR_PRIVATE_KEY_PATH",
		},
		// Role is the role that we expect the host to be (Validator, API, AWMRelayer, Loadtest or
		// Monitor)
		Roles: []node.SupportedRole{node.Validator},
	}

	// Here we are assuming that the node is currently validating the Primary Network, which is
	// a requirement before the node can start validating a Subnet.
	// To have a node validate the Primary Network, call node.ValidatePrimaryNetwork
	// Now we are calling the node to start tracking the Subnet
	subnetIDsToValidate := []string{newSubnet.SubnetID.String()}
	if err := node.SyncSubnets(subnetIDsToValidate); err != nil {
		panic(err)
	}

	// Node is now tracking the Subnet

	// Key that will be used for paying the transaction fees of Subnet AddValidator Tx
	//
	// In our example, this Key is also the control Key to the Subnet, so we are going to use
	// this key to also sign the Subnet AddValidator tx
	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "PRIVATE_KEY_FILEPATH", nil)
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

	nodeID, err := ids.NodeIDFromString(node.NodeID)
	if err != nil {
		panic(err)
	}

	validatorParams := validator.SubnetValidatorParams{
		NodeID: nodeID,
		// Validate Subnet for 48 hours
		Duration: 48 * time.Hour,
		Weight:   20,
	}

	// We need to set Subnet Auth Keys for this transaction since Subnet AddValidator is
	// a Subnet-changing transaction
	//
	// In this example, the example Subnet was created with only 1 key as control key with a threshold of 1
	// and the control key is the key contained in the keychain object, so we are going to use the
	// key contained in the keychain object as the Subnet Auth Key for Subnet AddValidator tx
	subnetAuthKeys := keychain.Addresses().List()
	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)

	addValidatorTx, err := newSubnet.AddValidator(wallet, validatorParams)
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
