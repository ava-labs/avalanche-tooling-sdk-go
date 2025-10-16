// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package validatormanager

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
)

// PoAValidatorManagerInitialize initializes contract [managerAddress] at [client], to
// manage validators on [subnetID], with
// owner given by [ownerAddress]
func PoAValidatorManagerInitialize(
	logger logging.Logger,
	client evm.Client,
	managerAddress common.Address,
	signer *evm.Signer,
	subnetID ids.ID,
	ownerAddress common.Address,
	useACP99 bool,
) (*types.Transaction, *types.Receipt, error) {
	if useACP99 {
		return client.TxToMethod(
			logger,
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
	return client.TxToMethod(
		logger,
		signer,
		managerAddress,
		nil,
		"initialize PoA manager",
		ErrorSignatureToError,
		"initialize((bytes32,uint64,uint8),address)",
		ValidatorManagerSettings{
			SubnetID:               subnetID,
			ChurnPeriodSeconds:     ChurnPeriodSeconds,
			MaximumChurnPercentage: MaximumChurnPercentage,
		},
		ownerAddress,
	)
}
