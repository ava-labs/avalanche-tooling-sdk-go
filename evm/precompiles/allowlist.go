// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package precompiles

import (
	"math/big"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
)

// SetAdmin sets an address as an admin in the precompile allowlist at [precompile]
// using [client] and [signer]. The transaction is logged using [logger].
func SetAdmin(
	logger logging.Logger,
	client evm.Client,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := client.TxToMethod(
		logger,
		signer,
		precompile,
		nil,
		"set precompile admin",
		nil,
		"setAdmin(address)",
		toSet,
	)
	return err
}

// SetManager sets an address as a manager in the precompile allowlist at [precompile]
// using [client] and [signer]. The transaction is logged using [logger].
func SetManager(
	logger logging.Logger,
	client evm.Client,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := client.TxToMethod(
		logger,
		signer,
		precompile,
		nil,
		"set precompile manager",
		nil,
		"setManager(address)",
		toSet,
	)
	return err
}

// SetEnabled sets an address as enabled in the precompile allowlist at [precompile]
// using [client] and [signer]. The transaction is logged using [logger].
func SetEnabled(
	logger logging.Logger,
	client evm.Client,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := client.TxToMethod(
		logger,
		signer,
		precompile,
		nil,
		"set precompile enabled",
		nil,
		"setEnabled(address)",
		toSet,
	)
	return err
}

// SetNone removes an address from all roles in the precompile allowlist at [precompile]
// using [client] and [signer]. The transaction is logged using [logger].
func SetNone(
	logger logging.Logger,
	client evm.Client,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := client.TxToMethod(
		logger,
		signer,
		precompile,
		nil,
		"set precompile none",
		nil,
		"setNone(address)",
		toSet,
	)
	return err
}

// ReadAllowList queries the role of an address in the precompile allowlist at [precompile]
// using [client]. Returns the role as a big.Int (0=None, 1=Enabled, 2=Manager, 3=Admin).
func ReadAllowList(
	client evm.Client,
	precompile common.Address,
	toQuery common.Address,
) (*big.Int, error) {
	out, err := client.CallToMethod(
		precompile,
		"readAllowList(address)->(uint256)",
		nil,
		toQuery,
	)
	if err != nil {
		return nil, err
	}
	return evm.GetSmartContractCallResult[*big.Int]("readAllowList", out)
}
