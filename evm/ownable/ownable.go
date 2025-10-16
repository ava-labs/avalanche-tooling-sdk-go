// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package ownable provides utilities for interacting with OpenZeppelin Ownable contracts.
package ownable

import (
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
)

// GetContractOwner retrieves the owner address of an OpenZeppelin Ownable contract at [contractAddress] using [client].
// See https://docs.openzeppelin.com/contracts/2.x/api/ownership#Ownable-owner for contract details.
func GetContractOwner(
	client evm.Client,
	contractAddress common.Address,
) (common.Address, error) {
	out, err := client.CallToMethod(
		contractAddress,
		"owner()->(address)",
		nil,
	)
	if err != nil {
		return common.Address{}, err
	}
	return evm.GetSmartContractCallResult[common.Address]("owner", out)
}

// TransferOwnership transfers ownership of an OpenZeppelin Ownable contract at [contractAddress]
// to [newOwner] using [client] and [signer]. The transaction is logged using [logger].
func TransferOwnership(
	logger logging.Logger,
	client evm.Client,
	contractAddress common.Address,
	signer *evm.Signer,
	newOwner common.Address,
) error {
	_, _, err := client.TxToMethod(
		logger,
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
