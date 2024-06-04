// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"fmt"
	"testing"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"golang.org/x/net/context"
)

func TestAddValidatorDeploy(_ *testing.T) {
	// Initialize a new Avalanche Object which will be used to set shared properties
	// like logging, metrics preferences, etc
	baseApp := avalanche.New(avalanche.DefaultLeveledLogger)
	subnetParams := SubnetParams{
		SubnetEVM: SubnetEVMParams{
			EvmChainID:       1234567,
			EvmDefaults:      true,
			EnableWarp:       true,
			EnableTeleporter: true,
			EnableRelayer:    true,
		},
	}
	newSubnet, _ := New(baseApp, &subnetParams)
	ctx := context.Background()
	wallet, _ := wallet.New(
		ctx,
		&primary.WalletConfig{
			URI:              "",
			AVAXKeychain:     nil,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	// deploy Subnet returns multisig and error
	createSubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	fmt.Printf("deploySubnetTx %s", createSubnetTx)
	createBlockchainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	fmt.Printf("deployBlockchainTx %s", createBlockchainTx)
	nodeID, _ := ids.NodeIDFromString("node-123")
	validatorInput := ValidatorParams{
		NodeID:   nodeID,
		Duration: time.Hour * 48,
		Weight:   20,
		Network:  avalanche.Network{Kind: avalanche.Fuji},
	}
	addValidatorTx, _ := newSubnet.AddValidator(wallet, validatorInput)
	fmt.Printf("addValidatorTx %s", addValidatorTx)
}
