// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package precompiles

import (
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
)

func WarpPrecompileGetBlockchainID(
	client evm.Client,
) (ids.ID, error) {
	out, err := client.CallToMethod(
		WarpPrecompile,
		"getBlockchainID()->(bytes32)",
		nil,
	)
	if err != nil {
		return ids.Empty, err
	}
	return evm.GetSmartContractCallResult[[32]byte]("getBlockchainID", out)
}
