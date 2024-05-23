// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"avalanche-tooling-sdk-go/avalanche"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/coreth/core"
	"github.com/ava-labs/coreth/utils"
	"github.com/ava-labs/subnet-evm/params"
	"math/big"
	"time"
)

type SubnetParams struct {
	// File path of Genesis to use
	// Do not set EvmChainID, EvmToken and EvmDefaults values in SubnetEVM
	// if GenesisFilePath value is set

	// See https://docs.avax.network/build/subnet/upgrade/customize-a-subnet#genesis for
	// information on Genesis
	GenesisFilePath string

	// Subnet-EVM parameters to use
	// Do not set SubnetEVM value if you are using Custom VM
	SubnetEVM SubnetEVMParams

	// Custom VM parameters to use
	// Do not set CustomVM value if you are using Subnet-EVM
	CustomVM CustomVMParams

	Name string
}

type SubnetEVMParams struct {
	// Version of Subnet-EVM to use
	// Do not set EvmVersion value if UseLatestReleasedEvmVersion or
	// UseLatestPreReleasedEvmVersion is set to true

	// Available Subnet-EVM versions can be found at https://github.com/ava-labs/subnet-evm/releases
	EvmVersion string

	// Chain ID to use in Subnet-EVM
	EvmChainID uint64

	// Token name to use in Subnet-EVM
	EvmToken string

	// Use default settings for fees, airdrop, precompiles and teleporter in Subnet-EVM
	EvmDefaults bool

	// Use latest Subnet-EVM pre-released version
	// Available Subnet-EVM versions can be found at https://github.com/ava-labs/subnet-evm/releases
	UseLatestPreReleasedEvmVersion bool

	// Use latest Subnet-EVM version
	// Available Subnet-EVM versions can be found at https://github.com/ava-labs/subnet-evm/releases
	UseLatestReleasedEvmVersion bool

	// Enable Avalanche Warp Messaging (AWM) when deploying a VM

	// See https://docs.avax.network/build/cross-chain/awm/overview for
	// information on Avalanche Warp Messaging
	EnableWarp bool

	// Enable Teleporter when deploying a VM
	// Warp is required to be enabled when enabling Teleporter

	// See https://docs.avax.network/build/cross-chain/teleporter/overview for
	// information on Teleporter
	EnableTeleporter bool

	// Enable AWM Relayer when deploying a VM

	// See https://docs.avax.network/build/cross-chain/awm/relayer for
	// information on AWM Relayer
	EnableRelayer bool
}

type CustomVMParams struct {
	// File path of the Custom VM binary to use
	VMFilePath string

	// Git Repo URL to be used to build Custom VM
	// Only set CustomVMRepoURL value when VMFilePath value is not set
	CustomVMRepoURL string

	// Git branch or commit to be used to build Custom VM
	// Only set CustomVMBranch value when VMFilePath value is not set
	CustomVMBranch string

	// Filepath of the script to be used to build Custom VM
	// Only set CustomVMBuildScript value when VMFilePath value is not set
	CustomVMBuildScript string
}

type Subnet struct {
	Name string

	Genesis []byte

	ControlKeys []string

	SubnetAuthKeys []ids.ShortID

	SubnetID ids.ID

	TransferSubnetOwnershipTxID ids.ID

	Chain string

	Threshold uint32

	VMID ids.ID

	RPCVersion int

	TokenName string

	TokenSymbol string

	Logger avalanche.LeveledLoggerInterface
}

func New(client *avalanche.BaseApp, subnetParams *SubnetParams) *Subnet {
	subnet := Subnet{
		Logger: client.Logger,
	}
	return &subnet
}

func createEvmGenesis(
	subnetName string,
	subnetEVMVersion string,
	rpcVersion int,
	subnetEVMChainID uint64,
	subnetEVMTokenSymbol string,
	useSubnetEVMDefaults bool,
	useWarp bool,
	teleporterInfo *teleporter.Info,
) ([]byte, error) {
	genesis := core.Genesis{}
	genesis.Timestamp = *utils.TimeToNewUint64(time.Now())

	conf := params.SubnetEVMDefaultChainConfig
	conf.NetworkUpgrades = params.NetworkUpgrades{}

	const (
		descriptorsState = "descriptors"
		feeState         = "fee"
		airdropState     = "airdrop"
		precompilesState = "precompiles"
	)

	var (
		chainID     *big.Int
		tokenSymbol string
		allocation  core.GenesisAlloc
		//direction   statemachine.StateDirection
		err error
	)

	//subnetEvmState, err := statemachine.NewStateMachine(
	//	[]string{descriptorsState, feeState, airdropState, precompilesState},
	//)
	//if err != nil {
	//	return nil, err
	//}
	//for subnetEvmState.Running() {
	//	switch subnetEvmState.CurrentState() {
	//	case descriptorsState:
	//		chainID, tokenSymbol, direction, err = getDescriptors(
	//			app,
	//			subnetEVMChainID,
	//			subnetEVMTokenSymbol,
	//		)
	//	case feeState:
	//		*conf, direction, err = GetFeeConfig(*conf, app, useSubnetEVMDefaults)
	//	case airdropState:
	//		allocation, direction, err = getAllocation(
	//			app,
	//			subnetName,
	//			defaultEvmAirdropAmount,
	//			oneAvax,
	//			fmt.Sprintf("Amount to airdrop (in %s units)", tokenSymbol),
	//			useSubnetEVMDefaults,
	//		)
	//		if teleporterInfo != nil {
	//			allocation = addTeleporterAddressToAllocations(
	//				allocation,
	//				teleporterInfo.FundedAddress,
	//				teleporterInfo.FundedBalance,
	//			)
	//		}
	//	case precompilesState:
	//		*conf, direction, err = getPrecompiles(*conf, app, &genesis.Timestamp, useSubnetEVMDefaults, useWarp)
	//		if teleporterInfo != nil {
	//			*conf = addTeleporterAddressesToAllowLists(
	//				*conf,
	//				teleporterInfo.FundedAddress,
	//				teleporterInfo.MessengerDeployerAddress,
	//				teleporterInfo.RelayerAddress,
	//			)
	//		}
	//	default:
	//		err = errors.New("invalid creation stage")
	//	}
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	subnetEvmState.NextState(direction)
	//}

	if conf != nil && conf.GenesisPrecompiles[txallowlist.ConfigKey] != nil {
		allowListCfg, ok := conf.GenesisPrecompiles[txallowlist.ConfigKey].(*txallowlist.Config)
		if !ok {
			return nil, fmt.Errorf(
				"expected config of type txallowlist.AllowListConfig, but got %T",
				allowListCfg,
			)
		}

		if err := ensureAdminsHaveBalance(
			allowListCfg.AdminAddresses,
			allocation); err != nil {
			return nil, nil, err
		}
	}

	conf.ChainID = chainID

	genesis.Alloc = allocation
	genesis.Config = conf
	genesis.Difficulty = Difficulty
	genesis.GasLimit = conf.FeeConfig.GasLimit.Uint64()

	jsonBytes, err := genesis.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonBytes, "", "    ")
	if err != nil {
		return nil, err
	}

	//sc := &models.Sidecar{
	//	Name:        subnetName,
	//	VM:          models.SubnetEvm,
	//	VMVersion:   subnetEVMVersion,
	//	RPCVersion:  rpcVersion,
	//	Subnet:      subnetName,
	//	TokenSymbol: tokenSymbol,
	//	TokenName:   tokenSymbol + " Token",
	//}

	//return prettyJSON.Bytes(), sc, nil
	return prettyJSON.Bytes(), nil
}
