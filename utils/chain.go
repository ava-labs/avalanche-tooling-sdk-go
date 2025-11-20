// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/ids"
)

// GetChainID returns the blockchain ID for a given chain name (e.g., "C", "X", "P")
func GetChainID(endpoint string, chainName string) (ids.ID, error) {
	client := info.NewClient(endpoint)
	ctx, cancel := GetAPIContext()
	defer cancel()
	return client.GetBlockchainID(ctx, chainName)
}
