// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"avalanche-tooling-sdk-go/avalanche"
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"testing"
)

func TestSubnetDeploy(t *testing.T) {
	baseApp := avalanche.New()
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
	wallet, _ := primary.MakeWallet(
		ctx,
		&primary.WalletConfig{
			URI:              "",
			AVAXKeychain:     nil,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	//deploy Subnet returns tx ID and error
	deploySubnetTx, _ := createSubnetTx(newSubnet, wallet)
	fmt.Printf("deploySubnetTx %s", deploySubnetTx)
}
