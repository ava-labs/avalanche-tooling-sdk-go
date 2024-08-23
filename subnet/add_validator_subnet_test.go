// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func TestValidateSubnet(t *testing.T) {
	require := require.New(t)
	subnetParams := SubnetParams{
		GenesisFilePath: "GENESIS_FILE_PATH",
		Name:            "SUBNET_NAME",
	}

	newSubnet, err := New(&subnetParams)
	require.NoError(err)

	// Genesis doesn't contain the deployed Subnet's SubnetID, we need to first set the Subnet ID
	subnetID, err := ids.FromString("SUBNET_ID")
	require.NoError(err)

	newSubnet.SetSubnetID(subnetID)

	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "PRIVATE_KEY_FILEPATH", nil)
	require.NoError(err)

	wallet, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	require.NoError(err)

	nodeID, err := ids.NodeIDFromString("VALIDATOR_NODEID")
	require.NoError(err)

	validator := validator.SubnetValidatorParams{
		NodeID: nodeID,
		// Validate Subnet for 48 hours
		Duration: 48 * time.Hour,
		// Setting weight of subnet validator to 20 (default value)
		Weight: 20,
	}

	// We need to set Subnet Auth Keys for this transaction since Subnet AddValidator is
	// a Subnet-changing transaction
	//
	// In this example, the example Subnet was created with only 1 key as control key with a threshold of 1
	// and the control key is the key contained in the keychain object, so we are going to use the
	// key contained in the keychain object as the Subnet Auth Key for Subnet AddValidator tx
	subnetAuthKeys := keychain.Addresses().List()
	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)

	addValidatorTx, err := newSubnet.AddValidator(wallet, validator)
	require.NoError(err)

	// Since it has the required signatures, we will now commit the transaction on chain
	txID, err := newSubnet.Commit(*addValidatorTx, wallet, true)
	require.NoError(err)

	fmt.Printf("obtained tx id %s", txID.String())
}
