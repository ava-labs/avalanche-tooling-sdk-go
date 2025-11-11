// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package contract

import (
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
)

// GetContractOwner gets owner for https://docs.openzeppelin.com/contracts/2.x/api/ownership#Ownable-owner contracts
func GetContractOwner(
	rpcURL string,
	contractAddress common.Address,
) (common.Address, error) {
	out, err := CallToMethod(
		rpcURL,
		common.Address{},
		contractAddress,
		"owner()->(address)",
		nil,
	)
	if err != nil {
		return common.Address{}, err
	}
	return GetSmartContractCallResult[common.Address]("owner", out)
}

func TransferOwnership(
	logger logging.Logger,
	rpcURL string,
	contractAddress common.Address,
	signer *evm.Signer,
	newOwner common.Address,
) error {
	_, _, err := TxToMethod(
		logger,
		rpcURL,
		signer,
		contractAddress,
		nil,
		"transfer ownership",
		nil,
		"transferOwnership(address)",
		newOwner,
	)
	return err
}
