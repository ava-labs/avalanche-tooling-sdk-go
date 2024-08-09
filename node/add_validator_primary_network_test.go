// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"testing"
	"time"
)

func TestNodesValidatePrimaryNetwork(_ *testing.T) {
	ctx := context.Background()
	cp, err := GetDefaultCloudParams(ctx, AWSCloud)
	if err != nil {
		panic(err)
	}

	node := Node{
		// NodeID is Avalanche Node ID of the node
		NodeID: "NodeID-Mb3AwcUpWysCWLP6mSpzzJVgYawJWzPHu",
		// IP address of the node
		IP: "18.144.79.215",
		// SSH configuration for the node
		SSHConfig: SSHConfig{
			User:           constants.AnsibleSSHUser,
			PrivateKeyPath: "/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem",
		},
		// Cloud is the cloud service that the node is on
		Cloud: AWSCloud,
		// CloudConfig is the cloud specific configuration for the node
		CloudConfig: *cp,
		// Role of the node can be 	Validator, API, AWMRelayer, Loadtest, or Monitor
		Roles: []SupportedRole{Validator},
	}

	//err = node.ProvideStakingCertAndKey(fmt.Sprintf("/Users/raymondsukanto/.avalanche-cli/nodes/%s", node.NodeID))
	//if err != nil {
	//	panic(err)
	//}

	nodeID, err := ids.NodeIDFromString(node.NodeID)
	if err != nil {
		panic(err)
	}

	validator := ValidatorParams{
		NodeID: nodeID,
		// Validate Primary Network for 48 hours
		Duration: 48 * time.Hour,
		// 1 billion in weight is equivalent to 1 AVAX
		Weight: 2000000000,
	}

	network := avalanche.FujiNetwork()

	keychain, err := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk", nil)
	if err != nil {
		panic(err)
	}

	wallet, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	if err != nil {
		panic(err)
	}

	err = node.SetNodeBLSKey(fmt.Sprintf("/Users/raymondsukanto/.avalanche-cli/nodes/%s/signer.key", node.NodeID))
	if err != nil {
		panic(err)
	}

	txID, err := node.AddNodeAsPrimaryNetworkValidator(avalanche.FujiNetwork(), validator, wallet)
	if err != nil {
		panic(err)
	}
	fmt.Printf("obtained tx id %s", txID.String())
}
