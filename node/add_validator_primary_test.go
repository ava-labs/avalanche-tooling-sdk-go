// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func TestNodesValidatePrimaryNetwork(t *testing.T) {
	require := require.New(t)
	// We are using an existing host
	node := Node{
		// NodeID is Avalanche Node ID of the node
		NodeID: "NODE_ID",
		// IP address of the node
		IP: "NODE_IP_ADDRESS",
		// SSH configuration for the node
		SSHConfig: SSHConfig{
			User:           constants.RemoteHostUser,
			PrivateKeyPath: "NODE_KEYPAIR_PRIVATE_KEY_PATH",
		},
		// Role of the node can be 	Validator, API, AWMRelayer, Loadtest, or Monitor
		Roles: []SupportedRole{Validator},
	}

	nodeID, err := ids.NodeIDFromString(node.NodeID)
	require.NoError(err)

	validatorParams := validator.PrimaryNetworkValidatorParams{
		NodeID: nodeID,
		// Validate Primary Network for 48 hours
		Duration: 48 * time.Hour,
		// Stake 2 AVAX
		StakeAmount: 2 * units.Avax,
	}

	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "PRIVATE_KEY_FILEPATH", nil)
	require.NoError(err)

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

	txID, err := node.ValidatePrimaryNetwork(avalanche.FujiNetwork(), validatorParams, wallet)
	require.NoError(err)

	fmt.Printf("obtained tx id %s", txID.String())
}
