// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/subnet-evm/commontype"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ava-labs/subnet-evm/params/extras"
)

type SubnetEVMParams struct {
	// ChainID identifies the current chain and is used for replay protection
	ChainID *big.Int

	// FeeConfig sets the configuration for the dynamic fee algorithm
	FeeConfig commontype.FeeConfig

	// Allocation specifies the initial state that is part of the genesis block.
	Allocation core.GenesisAlloc

	// Ethereum uses Precompiles to efficiently implement cryptographic primitives within the EVM
	// instead of re-implementing the same primitives in Solidity.
	//
	// Precompiles are a shortcut to execute a function implemented by the EVM itself,
	// rather than an actual contract. A precompile is associated with a fixed address defined in
	// the EVM. There is no byte code associated with that address.
	//
	// For more information regarding Precompiles, head to https://docs.avax.network/build/vm/evm/intro.
	Precompiles extras.Precompiles

	// Timestamp
	// TODO: add description what timestamp is
	Timestamp *uint64
}

func CreateEvmGenesis(
	subnetEVMParams *SubnetEVMParams,
) ([]byte, error) {
	genesis := core.Genesis{}
	genesis.Timestamp = *subnetEVMParams.Timestamp

	conf := params.SubnetEVMDefaultChainConfig
	extra := params.GetExtra(conf)

	extra.NetworkUpgrades = extras.NetworkUpgrades{}

	var err error

	if subnetEVMParams.ChainID == nil {
		return nil, fmt.Errorf("genesis params chain ID cannot be empty")
	}

	if subnetEVMParams.FeeConfig == commontype.EmptyFeeConfig {
		return nil, fmt.Errorf("genesis params fee config cannot be empty")
	}

	if subnetEVMParams.Allocation == nil {
		return nil, fmt.Errorf("genesis params allocation cannot be empty")
	}
	allocation := subnetEVMParams.Allocation

	if subnetEVMParams.Precompiles == nil {
		return nil, fmt.Errorf("genesis params precompiles cannot be empty")
	}

	extra.FeeConfig = subnetEVMParams.FeeConfig
	extra.GenesisPrecompiles = subnetEVMParams.Precompiles

	conf.ChainID = subnetEVMParams.ChainID

	genesis.Alloc = allocation
	genesis.Config = conf
	genesis.Difficulty = vm.Difficulty
	genesis.GasLimit = extra.FeeConfig.GasLimit.Uint64()

	jsonBytes, err := genesis.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonBytes, "", "    ")
	if err != nil {
		return nil, err
	}

	return prettyJSON.Bytes(), nil
}

func GetDefaultSubnetEVMGenesis(initialAllocationAddress string) SubnetEVMParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress(initialAllocationAddress)] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	return SubnetEVMParams{
		ChainID:     big.NewInt(123456),
		FeeConfig:   vm.StarterFeeConfig,
		Allocation:  allocation,
		Precompiles: extras.Precompiles{},
	}
}

func VmID(vmName string) (ids.ID, error) {
	if len(vmName) > 32 {
		return ids.Empty, fmt.Errorf("VM name must be <= 32 bytes, found %d", len(vmName))
	}
	b := make([]byte, 32)
	copy(b, []byte(vmName))
	return ids.ToID(b)
}
