// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"avalanche-tooling-sdk-go/avalanche"
	"context"
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func TestSubnetDeploy(t *testing.T) {
	baseApp := avalanche.New(avalanche.DefaultLeveledLogger)
	subnetParams := SubnetParams{
		SubnetEVM: SubnetEVMParams{
			EvmChainID:                  1234567,
			EvmToken:                    "AVAX",
			EvmDefaults:                 true,
			UseLatestReleasedEvmVersion: true,
			EnableWarp:                  true,
			EnableTeleporter:            true,
			EnableRelayer:               true,
		},
	}
	newSubnet := New(baseApp, &subnetParams)
	ctx := context.Background()
	wallet, _ := NewWallet(
		ctx,
		&primary.WalletConfig{
			URI:              "",
			AVAXKeychain:     nil,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	//deploy Subnet returns tx ID and error
	deploySubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	fmt.Printf("deploySubnetTx %s", deploySubnetTx)
}
