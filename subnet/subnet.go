// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ava-labs/subnet-evm/precompile/contracts/warp"

	"github.com/ava-labs/avalanche-tooling-sdk-go/teleporter"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/coreth/utils"
	"github.com/ava-labs/subnet-evm/commontype"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ava-labs/subnet-evm/precompile/contracts/txallowlist"
	"github.com/ethereum/go-ethereum/common"
)

type SubnetParams struct {
	// File path of Genesis to use
	// Do not set SubnetEVMParams or CustomVMParams
	// if GenesisFilePath value is set

	// See https://docs.avax.network/build/subnet/upgrade/customize-a-subnet#genesis for
	// information on Genesis
	GenesisFilePath string

	// Subnet-EVM parameters to use
	// Do not set SubnetEVM value if you are using Custom VM
	SubnetEVM *SubnetEVMParams

	// Custom VM parameters to use
	// Do not set CustomVM value if you are using Subnet-EVM
	CustomVM *CustomVMParams

	Name string
}

type SubnetEVMParams struct {
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

	GenesisParams *EVMGenesisParams
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

	SubnetID ids.ID

	VMID ids.ID

	DeployInfo DeployParams

	RPCVersion int
}

func (c *Subnet) SetDeployParams(controlKeys []string, subnetAuthKeys []ids.ShortID, threshold uint32) {
	c.DeployInfo = DeployParams{
		ControlKeys:    controlKeys,
		SubnetAuthKeys: subnetAuthKeys,
		Threshold:      threshold,
	}
}

type DeployParams struct {
	ControlKeys []string

	SubnetAuthKeys []ids.ShortID

	TransferSubnetOwnershipTxID ids.ID

	Threshold uint32
}

type EVMGenesisParams struct {
	// Chain ID to use in Subnet-EVM
	ChainID        *big.Int
	FeeConfig      commontype.FeeConfig
	Allocation     core.GenesisAlloc
	Precompiles    params.Precompiles
	TeleporterInfo *teleporter.Info
}

func New(subnetParams *SubnetParams) (*Subnet, error) {
	if subnetParams.GenesisFilePath != "" && (subnetParams.CustomVM != nil || subnetParams.SubnetEVM != nil) {
		return nil, fmt.Errorf("genesis file path cannot be non-empty if either CustomVM params or SubnetEVM params is not empty")
	}
	if subnetParams.SubnetEVM == nil && subnetParams.CustomVM != nil {
		return nil, fmt.Errorf("SubnetEVM params and CustomVM params cannot both be non-empty")
	}
	if subnetParams.SubnetEVM != nil {
		if subnetParams.SubnetEVM.GenesisParams == nil {
			return nil, fmt.Errorf("SubnetEVM Genesis params cannot be empty")
		}
	}
	var genesisBytes []byte
	var err error
	switch {
	case subnetParams.GenesisFilePath != "":
		genesisBytes, err = os.ReadFile(subnetParams.GenesisFilePath)
	case subnetParams.SubnetEVM != nil:
		genesisBytes, err = createEvmGenesis(
			subnetParams.SubnetEVM.GenesisParams,
			subnetParams.SubnetEVM.EnableWarp,
		)
	case subnetParams.CustomVM != nil:
		genesisBytes, err = createCustomVMGenesis()
	default:
	}
	if err != nil {
		return nil, err
	}
	vmID, err := vmID(subnetParams.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM ID from %s: %w", subnetParams.Name, err)
	}
	subnet := Subnet{
		Name:    subnetParams.Name,
		VMID:    vmID,
		Genesis: genesisBytes,
	}
	return &subnet, nil
}

func createEvmGenesis(
	genesisParams *EVMGenesisParams,
	useWarp bool,
) ([]byte, error) {
	genesis := core.Genesis{}
	genesis.Timestamp = *utils.TimeToNewUint64(time.Now())

	conf := params.SubnetEVMDefaultChainConfig
	conf.NetworkUpgrades = params.NetworkUpgrades{}

	var err error

	if genesisParams.ChainID == nil {
		return nil, fmt.Errorf("genesis params chain ID cannot be empty")
	}

	if genesisParams.FeeConfig == commontype.EmptyFeeConfig {
		return nil, fmt.Errorf("genesis params fee config cannot be empty")
	}

	if genesisParams.Allocation == nil {
		return nil, fmt.Errorf("genesis params allocation cannot be empty")
	}
	allocation := genesisParams.Allocation
	if genesisParams.TeleporterInfo != nil {
		allocation = addTeleporterAddressToAllocations(
			allocation,
			genesisParams.TeleporterInfo.FundedAddress,
			genesisParams.TeleporterInfo.FundedBalance,
		)
	}

	if genesisParams.Precompiles == nil {
		return nil, fmt.Errorf("genesis params precompiles cannot be empty")
	}
	if useWarp {
		warpConfig := vm.ConfigureWarp(&genesis.Timestamp)
		conf.GenesisPrecompiles[warp.ConfigKey] = &warpConfig
	}
	if genesisParams.TeleporterInfo != nil {
		*conf = vm.AddTeleporterAddressesToAllowLists(
			*conf,
			genesisParams.TeleporterInfo.FundedAddress,
			genesisParams.TeleporterInfo.MessengerDeployerAddress,
			genesisParams.TeleporterInfo.RelayerAddress,
		)
	}

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
			return nil, err
		}
	}

	conf.ChainID = genesisParams.ChainID

	genesis.Alloc = allocation
	genesis.Config = conf
	genesis.Difficulty = vm.Difficulty
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

	return prettyJSON.Bytes(), nil
}

func ensureAdminsHaveBalance(admins []common.Address, alloc core.GenesisAlloc) error {
	if len(admins) < 1 {
		return nil
	}

	for _, admin := range admins {
		// we can break at the first admin who has a non-zero balance
		if bal, ok := alloc[admin]; ok &&
			bal.Balance != nil &&
			bal.Balance.Uint64() > uint64(0) {
			return nil
		}
	}
	return errors.New(
		"none of the addresses in the transaction allow list precompile have any tokens allocated to them. Currently, no address can transact on the network. Airdrop some funds to one of the allow list addresses to continue",
	)
}

func addAllocation(alloc core.GenesisAlloc, address string, amount *big.Int) {
	alloc[common.HexToAddress(address)] = core.GenesisAccount{
		Balance: amount,
	}
}

func addTeleporterAddressToAllocations(
	alloc core.GenesisAlloc,
	teleporterKeyAddress string,
	teleporterKeyBalance *big.Int,
) core.GenesisAlloc {
	if alloc != nil {
		addAllocation(alloc, teleporterKeyAddress, teleporterKeyBalance)
	}
	return alloc
}

// TODO: implement createCustomVMGenesis
func createCustomVMGenesis() ([]byte, error) {
	return nil, nil
}

func vmID(vmName string) (ids.ID, error) {
	if len(vmName) > 32 {
		return ids.Empty, fmt.Errorf("VM name must be <= 32 bytes, found %d", len(vmName))
	}
	b := make([]byte, 32)
	copy(b, []byte(vmName))
	return ids.ToID(b)
}
