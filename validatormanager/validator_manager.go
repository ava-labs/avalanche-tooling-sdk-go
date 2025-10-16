// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package validatormanager

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/subnet-evm/accounts/abi"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"
)

// GetValidatorReturn represents the data returned when querying validator information.
type GetValidatorReturn struct {
	Status         uint8
	NodeID         []byte
	StartingWeight uint64
	SentNonce      uint64
	ReceivedNonce  uint64
	Weight         uint64
	StartTime      uint64
	EndTime        uint64
}

// GetValidator retrieves validator information for the given validation ID from the validator manager contract at [managerAddress] using [client].
func GetValidator(
	client evm.Client,
	managerAddress common.Address,
	validationID ids.ID,
) (*GetValidatorReturn, error) {
	stakingManagerSettings, err := GetStakingManagerSettings(
		client,
		managerAddress,
	)
	if err == nil {
		// fix address if specialized
		managerAddress = stakingManagerSettings.ValidatorManager
	}
	getValidatorReturn := &GetValidatorReturn{}
	out, err := client.CallToMethod(
		managerAddress,
		"getValidator(bytes32)->((uint8,bytes,uint64,uint64,uint64,uint64,uint64,uint64))",
		[]interface{}{*getValidatorReturn},
		[32]byte(validationID),
	)
	if err != nil {
		return getValidatorReturn, err
	}
	if len(out) != 1 {
		return getValidatorReturn, fmt.Errorf("incorrect number of outputs for getValidator: expected 1 got %d", len(out))
	}
	var ok bool
	getValidatorReturn, ok = abi.ConvertType(out[0], new(GetValidatorReturn)).(*GetValidatorReturn)
	if !ok {
		return getValidatorReturn, fmt.Errorf("invalid type for output of getValidator: expected GetValidatorReturn, got %T", out[0])
	}
	return getValidatorReturn, nil
}

// ChurnSettings contains the churn configuration for a validator manager.
type ChurnSettings struct {
	ChurnPeriodSeconds     uint64
	MaximumChurnPercentage uint8
}

// GetChurnSettings retrieves the churn tracker settings from the validator manager contract at [managerAddress] using [client].
func GetChurnSettings(
	client evm.Client,
	managerAddress common.Address,
) (ChurnSettings, error) {
	stakingManagerSettings, err := GetStakingManagerSettings(
		client,
		managerAddress,
	)
	if err == nil {
		// fix address if specialized
		managerAddress = stakingManagerSettings.ValidatorManager
	}
	churnSettings := ChurnSettings{}
	out, err := client.CallToMethod(
		managerAddress,
		"getChurnTracker()->(uint64,uint8,uint256,uint64,uint64,uint64)",
		nil,
	)
	if err != nil {
		return churnSettings, err
	}
	if len(out) != 6 {
		return churnSettings, fmt.Errorf("incorrect number of outputs for getChurnTracker: expected 6 got %d", len(out))
	}
	var ok bool
	churnSettings.ChurnPeriodSeconds, ok = out[0].(uint64)
	if !ok {
		return churnSettings, fmt.Errorf("invalid type for churnPeriodSeconds output of getChurnTracker: expected uint64, got %T", out[0])
	}
	churnSettings.MaximumChurnPercentage, ok = out[1].(uint8)
	if !ok {
		return churnSettings, fmt.Errorf("invalid type for maximumChurnPercentage output of getChurnTracker: expected uint8, got %T", out[1])
	}
	return churnSettings, nil
}

// Returns the validation ID for the Node ID, as registered at the validator manager
// Will return ids.Empty in case it is not registered
func GetValidationID(
	client evm.Client,
	managerAddress common.Address,
	nodeID ids.NodeID,
) (ids.ID, error) {
	stakingManagerSettings, err := GetStakingManagerSettings(
		client,
		managerAddress,
	)
	if err == nil {
		// fix address if specialized
		managerAddress = stakingManagerSettings.ValidatorManager
	}
	out, err := client.CallToMethod(
		managerAddress,
		"registeredValidators(bytes)->(bytes32)",
		nil,
		nodeID[:],
	)
	if err != nil {
		return ids.Empty, err
	}
	return evm.GetSmartContractCallResult[[32]byte]("registeredValidators", out)
}

// GetSubnetID retrieves the subnet ID associated with the validator manager contract at [managerAddress] using [client].
func GetSubnetID(
	client evm.Client,
	managerAddress common.Address,
) (ids.ID, error) {
	stakingManagerSettings, err := GetStakingManagerSettings(
		client,
		managerAddress,
	)
	if err == nil {
		// fix address if specialized
		managerAddress = stakingManagerSettings.ValidatorManager
	}
	out, err := client.CallToMethod(
		managerAddress,
		"subnetID()->(bytes32)",
		nil,
	)
	if err != nil {
		return ids.Empty, err
	}
	return evm.GetSmartContractCallResult[[32]byte]("subnetID", out)
}

// CurrentWeightInfo contains information about the current weight limits and churn settings for a validator manager.
type CurrentWeightInfo struct {
	TotalWeight       uint64
	MaximumPercentage uint8
	AllowedWeight     float64
}

// GetCurrentWeightInfo retrieves the current weight information including total weight and churn limits
// from the validator manager contract at [managerAddress] using [client] and [network].
func GetCurrentWeightInfo(
	network network.Network,
	client evm.Client,
	managerAddress common.Address,
) (CurrentWeightInfo, error) {
	subnetID, err := GetSubnetID(
		client,
		managerAddress,
	)
	if err != nil {
		return CurrentWeightInfo{}, err
	}
	totalWeight, err := validator.GetTotalWeight(network, subnetID)
	if err != nil {
		return CurrentWeightInfo{}, err
	}
	churnSettings, err := GetChurnSettings(
		client,
		managerAddress,
	)
	if err != nil {
		return CurrentWeightInfo{}, err
	}
	return CurrentWeightInfo{
		TotalWeight:       totalWeight,
		MaximumPercentage: churnSettings.MaximumChurnPercentage,
		AllowedWeight:     float64(totalWeight*uint64(churnSettings.MaximumChurnPercentage)) / 100.0,
	}, nil
}
