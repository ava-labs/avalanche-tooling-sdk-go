// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"context"
	"fmt"

	"github.com/ava-labs/subnet-evm/rpc"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// also used at mocks
var (
	rpcDialContext = rpc.DialContext
)

// wraps over rpc.Client for calls used by SDK. used to make evm calls not available in ethclient:
// - debug trace call
// - debug trace transaction
// features:
// - finds out url scheme in case it is missing, to connect to ws/wss/http/https
// - repeats to try to recover from failures, generating its own context for each call
// - logs rpc url in case of failure
type RawClient struct {
	RPCClient *rpc.Client
	URL       string
	// also used at mocks
	CallContext func(context.Context, interface{}, string, ...interface{}) error
}

// connects a raw evm rpc client to the given [rpcURL]
// supports [repeatsOnFailure] failures
func GetRawClient(rpcURL string) (RawClient, error) {
	client := RawClient{
		URL: rpcURL,
	}
	hasScheme, err := HasScheme(rpcURL)
	if err != nil {
		return RawClient{}, err
	}
	client.RPCClient, err = utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*rpc.Client, error) {
			if hasScheme {
				return rpcDialContext(ctx, rpcURL)
			} else {
				_, scheme, err := GetClientWithoutScheme(rpcURL)
				if err != nil {
					return nil, err
				}
				return rpcDialContext(ctx, scheme+rpcURL)
			}
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure connecting to rpc client on %s: %w", rpcURL, err)
	}
	client.CallContext = client.RPCClient.CallContext
	return client, err
}

// closes underlying rpc connection
func (client RawClient) Close() {
	client.RPCClient.Close()
}

// returns a trace for the given [txID] on [client]
// supports [repeatsOnFailure] failures
func (client RawClient) DebugTraceTransaction(
	txID string,
) (map[string]interface{}, error) {
	trace, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (map[string]interface{}, error) {
			var trace map[string]interface{}
			err := client.CallContext(
				ctx,
				&trace,
				"debug_traceTransaction",
				txID,
				map[string]string{"tracer": "callTracer"},
			)
			return trace, err
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure tracing tx %s on %s: %w", txID, client.URL, err)
	}
	return trace, err
}

// returns a trace for making a call on [client] with the given [data]
// supports [repeatsOnFailure] failures
func (client RawClient) DebugTraceCall(
	data map[string]string,
) (map[string]interface{}, error) {
	trace, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (map[string]interface{}, error) {
			var trace map[string]interface{}
			err := client.CallContext(
				ctx,
				&trace,
				"debug_traceCall",
				data,
				"latest",
				map[string]interface{}{
					"tracer": "callTracer",
					"tracerConfig": map[string]interface{}{
						"onlyTopCall": false,
					},
				},
			)
			return trace, err
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure tracing call on %s: %w", client.URL, err)
	}
	return trace, err
}

// returns a trace for the given [txID] on [rpcURL]
// supports [repeatsOnFailure] failures
func GetTxTrace(rpcURL string, txID string) (map[string]interface{}, error) {
	client, err := GetRawClient(rpcURL)
	if err != nil {
		return nil, err
	}
	return client.DebugTraceTransaction(txID)
}
