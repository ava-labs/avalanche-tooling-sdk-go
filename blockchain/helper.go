// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blockchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

func GetSubnet(subnetID ids.ID, network network.Network) (platformvm.GetSubnetClientResponse, error) {
	api := network.Endpoint
	pClient := platformvm.NewClient(api)
	ctx, cancel := utils.GetAPIContext()
	defer cancel()
	return pClient.GetSubnet(ctx, subnetID)
}
