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
	//
	// See https://docs.avax.network/build/subnet/upgrade/customize-a-subnet#genesis for
	// information on Genesis
	GenesisFilePath string

	// Subnet-EVM parameters to use
	// Do not set SubnetEVM value if you are using Custom VM
	SubnetEVM *SubnetEVMParams

	// Name is alias for the Subnet, it is used to derive VM ID, which is required
	// during for createBlockchainTx
	Name string
}

type SubnetEVMParams struct {
	// EnableWarp sets whether to enable Avalanche Warp Messaging (AWM) when deploying a VM
	//
	// See https://docs.avax.network/build/cross-chain/awm/overview for
	// information on Avalanche Warp Messaging
	EnableWarp bool

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
	Precompiles params.Precompiles

	// TeleporterInfo contains all the necessary information to dpeloy Teleporter into a Subnet
	//
	// If TeleporterInfo is not empty:
	// - Allocation will automatically be configured to add the provided Teleporter Address
	//   and Balance
	// - Precompiles tx allow list will include the provided Teleporter info
	TeleporterInfo *teleporter.Info
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
	// Name is alias for the Subnet
	Name string

	// Genesis is the initial state of a blockchain when it is first created. Each Virtual Machine
	// defines the format and semantics of its genesis data.
	//
	// For more information regarding Genesis, head to https://docs.avax.network/build/subnet/upgrade/customize-a-subnet#genesis
	Genesis []byte

	// SubnetID is the transaction ID from an issued CreateSubnetTX and is used to identify
	// the target Subnet for CreateChainTx and AddValidatorTx
	SubnetID ids.ID

	// VMID specifies the vm that the new chain will run when CreateChainTx is called
	VMID ids.ID

	// DeployInfo contains all the necessary information for createSubnetTx
	DeployInfo DeployParams
}

func (c *Subnet) SetDeployParams(controlKeys []ids.ShortID, subnetAuthKeys []ids.ShortID, threshold uint32) {
	c.DeployInfo = DeployParams{
		ControlKeys:    controlKeys,
		SubnetAuthKeys: subnetAuthKeys,
		Threshold:      threshold,
	}
}

type DeployParams struct {
	// ControlKeys is a list of P-Chain addresses that are authorized to create new chains and add
	// new validators to the Subnet
	ControlKeys []ids.ShortID

	// SubnetAuthKeys is a list of P-Chain addresses that will be used to sign transactions that
	// will modify the Subnet.
	//
	// SubnetAuthKeys has to be a subset of ControlKeys
	SubnetAuthKeys []ids.ShortID

	// Threshold is the minimum number of signatures needed before a transaction can be issued
	// Number of addresses in SubnetAuthKeys has to be more than or equal to Threshold number
	Threshold uint32
}

type EVMGenesisParams struct {
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
	Precompiles params.Precompiles

	// TeleporterInfo contains all the necessary information to dpeloy Teleporter into a Subnet
	//
	// If TeleporterInfo is not empty:
	// - Allocation will automatically be configured to add the provided Teleporter Address
	//   and Balance
	// - Precompiles tx allow list will include the provided Teleporter info
	TeleporterInfo *teleporter.Info
}

// New takes SubnetParams as input and creates Subnet as an output
//
// The created Subnet object can be used to :
//   - Create the Subnet on a specified network (Fuji / Mainnet)
//   - Create Blockchain(s) in the Subnet
//   - Add Validator(s) into the Subnet
func New(subnetParams *SubnetParams) (*Subnet, error) {
	if subnetParams.GenesisFilePath != "" && subnetParams.SubnetEVM != nil {
		return nil, fmt.Errorf("genesis file path cannot be non-empty if SubnetEVM params is not empty")
	}

	if subnetParams.GenesisFilePath == "" && subnetParams.SubnetEVM == nil {
		return nil, fmt.Errorf("genesis file path and SubnetEVM params params cannot all be empty")
	}

	if subnetParams.Name == "" {
		return nil, fmt.Errorf("SubnetEVM name cannot be empty")
	}

	var genesisBytes []byte
	var err error
	switch {
	case subnetParams.GenesisFilePath != "":
		genesisBytes, err = os.ReadFile(subnetParams.GenesisFilePath)
	case subnetParams.SubnetEVM != nil:
		genesisBytes, err = createEvmGenesis(subnetParams.SubnetEVM)
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
	subnetEVMParams *SubnetEVMParams,
) ([]byte, error) {
	genesis := core.Genesis{}
	genesis.Timestamp = *utils.TimeToNewUint64(time.Now())

	conf := params.SubnetEVMDefaultChainConfig
	conf.NetworkUpgrades = params.NetworkUpgrades{}

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
	if subnetEVMParams.TeleporterInfo != nil {
		allocation = addTeleporterAddressToAllocations(
			allocation,
			subnetEVMParams.TeleporterInfo.FundedAddress,
			subnetEVMParams.TeleporterInfo.FundedBalance,
		)
	}

	if subnetEVMParams.Precompiles == nil {
		return nil, fmt.Errorf("genesis params precompiles cannot be empty")
	}
	if subnetEVMParams.EnableWarp {
		warpConfig := vm.ConfigureWarp(&genesis.Timestamp)
		conf.GenesisPrecompiles[warp.ConfigKey] = &warpConfig
	}
	if subnetEVMParams.TeleporterInfo != nil {
		*conf = vm.AddTeleporterAddressesToAllowLists(
			*conf,
			subnetEVMParams.TeleporterInfo.FundedAddress,
			subnetEVMParams.TeleporterInfo.MessengerDeployerAddress,
			subnetEVMParams.TeleporterInfo.RelayerAddress,
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

	conf.ChainID = subnetEVMParams.ChainID

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

func vmID(vmName string) (ids.ID, error) {
	if len(vmName) > 32 {
		return ids.Empty, fmt.Errorf("VM name must be <= 32 bytes, found %d", len(vmName))
	}
	b := make([]byte, 32)
	copy(b, []byte(vmName))
	return ids.ToID(b)
}
