// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/ava-labs/avalanchego/api/connectclient"
	"github.com/ava-labs/avalanchego/connectproto/pb/proposervm/proposervmconnect"
	"github.com/ava-labs/avalanchego/vms/proposervm"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"

	pbproposervm "github.com/ava-labs/avalanchego/connectproto/pb/proposervm"
)

// GetCurrentEpoch returns the current epoch for a chain using JSON-RPC.
// This works with public endpoints that support JSON-RPC.
func GetCurrentEpoch(endpoint string, chainAlias string) (block.Epoch, error) {
	client := proposervm.NewJSONRPCClient(endpoint, chainAlias)
	ctx, cancel := GetAPIContext()
	defer cancel()
	return client.GetCurrentEpoch(ctx)
}

// GetCurrentL1Epoch returns the current epoch for an L1 blockchain.
// It first attempts to use Connect/gRPC (for local nodes), and falls back to JSON-RPC
// (for public endpoints) if Connect fails.
func GetCurrentL1Epoch(rpcURL, chainAlias string) (block.Epoch, error) {
	endpoint, err := url.Parse(rpcURL)
	if err != nil {
		return block.Epoch{}, fmt.Errorf("failed to parse rpc endpoint %w", err)
	}
	baseURL := fmt.Sprintf("%s://%s", endpoint.Scheme, endpoint.Host)
	proposerClient := proposervmconnect.NewProposerVMClient(
		connectclient.New(),
		baseURL,
		connect.WithInterceptors(
			connectclient.SetRouteHeaderInterceptor{
				Route: []string{
					chainAlias,
					proposervm.HTTPHeaderRoute,
				},
			},
		),
	)
	ctx, cancel := GetAPIContext()
	defer cancel()
	response, err := proposerClient.GetCurrentEpoch(ctx, &connect.Request[pbproposervm.GetCurrentEpochRequest]{})
	if err != nil {
		// Fall back to JSON-RPC if Connect/gRPC fails (e.g., public endpoints)
		// Check if chainAlias is a C-Chain blockchain ID and convert to "C" alias
		cChainID, cErr := GetChainID(baseURL, "C")
		if cErr == nil && cChainID.String() == chainAlias {
			chainAlias = "C"
		}
		return GetCurrentEpoch(baseURL, chainAlias)
	}
	return block.Epoch{
		StartTime:    response.Msg.StartTime,
		Number:       response.Msg.Number,
		PChainHeight: response.Msg.PChainHeight,
	}, nil
}
