// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package validatormanager

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm/contract"
)

// PoAValidatorManagerInitialize initializes contract [managerAddress] at [rpcURL], to
// manage validators on [subnetID], with
// owner given by [ownerAddress]
func PoAValidatorManagerInitialize(
	logger logging.Logger,
	rpcURL string,
	managerAddress common.Address,
	signer *evm.Signer,
	subnetID ids.ID,
	ownerAddress common.Address,
) (*types.Transaction, *types.Receipt, error) {
	return contract.TxToMethod(
		logger,
		rpcURL,
		signer,
		managerAddress,
		nil,
		"initialize PoA manager",
		ErrorSignatureToError,
		"initialize((address, bytes32,uint64,uint8))",
		ACP99ValidatorManagerSettings{
			Admin:                  ownerAddress,
			SubnetID:               subnetID,
			ChurnPeriodSeconds:     ChurnPeriodSeconds,
			MaximumChurnPercentage: MaximumChurnPercentage,
		},
	)
}
