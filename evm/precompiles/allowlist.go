// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package precompiles

import (
	"math/big"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm/contract"
)

func SetAdmin(
	logger logging.Logger,
	rpcURL string,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := contract.TxToMethod(
		logger,
		rpcURL,
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

func SetManager(
	logger logging.Logger,
	rpcURL string,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := contract.TxToMethod(
		logger,
		rpcURL,
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

func SetEnabled(
	logger logging.Logger,
	rpcURL string,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := contract.TxToMethod(
		logger,
		rpcURL,
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

func SetNone(
	logger logging.Logger,
	rpcURL string,
	precompile common.Address,
	signer *evm.Signer,
	toSet common.Address,
) error {
	_, _, err := contract.TxToMethod(
		logger,
		rpcURL,
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

func ReadAllowList(
	rpcURL string,
	precompile common.Address,
	toQuery common.Address,
) (*big.Int, error) {
	out, err := contract.CallToMethod(
		rpcURL,
		common.Address{},
		precompile,
		"readAllowList(address)->(uint256)",
		nil,
		toQuery,
	)
	if err != nil {
		return nil, err
	}
	return contract.GetSmartContractCallResult[*big.Int]("readAllowList", out)
}
