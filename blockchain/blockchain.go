// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/subnet-evm/commontype"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ava-labs/subnet-evm/params/extras"
	"go.uber.org/zap"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/validatormanager"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"

	subnetevmutils "github.com/ava-labs/subnet-evm/utils"
)

var (
	errMissingNetwork                            = fmt.Errorf("missing Network")
	errMissingSubnetID                           = fmt.Errorf("missing Subnet ID")
	errMissingBootstrapValidators                = fmt.Errorf("missing bootstrap validators")
	errMissingValidatorManagerBlockchainID       = fmt.Errorf("missing Validator Manager Blockchain ID")
	errMissingValidatorManagerRPC                = fmt.Errorf("missing Validator Manager RPC URL")
	errMissingValidatorManagerAddress            = fmt.Errorf("missing Validator Manager Address")
	errMissingSpecializedValidatorManagerAddress = fmt.Errorf("missing Specialized Validator Manager Address")
	errMissingValidatorManagerOwnerAddress       = fmt.Errorf("missing Validator Manager Owner Address")
	errMissingValidatorManagerOwnerPrivateKey    = fmt.Errorf("missing Validator Manager Owner Private Key")
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

	// Network where the subnet was deployed
	Network network.Network

	// SubnetID is the transaction ID from an issued CreateSubnetTX and is used to identify
	// the target Subnet for CreateChainTx and AddValidatorTx
	SubnetID ids.ID

	// VMID specifies the vm that the new chain will run when CreateChainTx is called
	VMID ids.ID

	// DeployInfo contains all the necessary information for createSubnetTx
	DeployInfo DeployParams

	// SubnetID where the Validator Manager is deployed
	ValidatorManagerSubnetID ids.ID

	// BlockchainID where the Validator Manager is deployed
	ValidatorManagerBlockchainID ids.ID

	// RPC URL the Validator Manager can be reached at
	ValidatorManagerRPC string

	// Address of the Validator Manager
	ValidatorManagerAddress *common.Address

	// Address of the Specialized Validator Manager
	SpecializedValidatorManagerAddress *common.Address

	// Address of the owner of the Validator Manager Contract
	ValidatorManagerOwnerAddress *common.Address

	// Signer of the owner of the Validator Manager Contract
	ValidatorManagerOwnerSigner *evm.Signer

	// BootstrapValidators are bootstrap validators that are included in the ConvertSubnetToL1Tx call
	// that made Subnet a sovereign L1
	BootstrapValidators []*txs.ConvertSubnetToL1Validator
}

func (c *Subnet) SetParams(controlKeys []ids.ShortID, subnetAuthKeys []ids.ShortID, threshold uint32) {
	c.DeployInfo = DeployParams{
		ControlKeys:    controlKeys,
		SubnetAuthKeys: subnetAuthKeys,
		Threshold:      threshold,
	}
}

// SetSubnetControlParams sets:
//   - control keys, which are keys that are allowed to make changes to a Subnet
//   - threshold, which is the number of keys that need to sign a transaction that changes
//     a Subnet
func (c *Subnet) SetSubnetControlParams(controlKeys []ids.ShortID, threshold uint32) {
	c.DeployInfo.ControlKeys = controlKeys
	c.DeployInfo.Threshold = threshold
}

// SetSubnetAuthKeys sets subnetAuthKeys, which are keys that are being used to sign a transaction
// that changes a Subnet
func (c *Subnet) SetSubnetAuthKeys(subnetAuthKeys []ids.ShortID) {
	c.DeployInfo.SubnetAuthKeys = subnetAuthKeys
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
		genesisBytes, err = CreateEvmGenesis(subnetParams.SubnetEVM)
	default:
	}
	if err != nil {
		return nil, err
	}

	vmID, err := VMID(subnetParams.Name)
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

func (c *Subnet) SetSubnetID(subnetID ids.ID) {
	c.SubnetID = subnetID
}

func CreateEvmGenesis(
	subnetEVMParams *SubnetEVMParams,
) ([]byte, error) {
	genesis := core.Genesis{}
	genesis.Timestamp = *subnetEVMParams.Timestamp

	conf := params.SubnetEVMDefaultChainConfig

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

	conf.ChainID = subnetEVMParams.ChainID

	genesis.Alloc = allocation
	genesis.Config = conf
	genesis.Difficulty = vm.Difficulty
	genesis.GasLimit = subnetEVMParams.FeeConfig.GasLimit.Uint64()

	var jsonBytes []byte
	params.WithTempRegisteredExtras(func() {
		chainExtras := *extras.SubnetEVMDefaultChainConfig
		chainExtras.FeeConfig = subnetEVMParams.FeeConfig
		chainExtras.GenesisPrecompiles = subnetEVMParams.Precompiles
		chainExtras.NetworkUpgrades = extras.NetworkUpgrades{}
		params.WithExtra(conf, &chainExtras)

		jsonBytes, err = genesis.MarshalJSON()
	})
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
	genesisBlock0Timestamp := subnetevmutils.TimeToNewUint64(time.Now())
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
		Timestamp:   genesisBlock0Timestamp,
	}
}

func VMID(vmName string) (ids.ID, error) {
	if len(vmName) > 32 {
		return ids.Empty, fmt.Errorf("VM name must be <= 32 bytes, found %d", len(vmName))
	}
	b := make([]byte, 32)
	copy(b, []byte(vmName))
	return ids.ToID(b)
}

// InitializeProofOfAuthority setups PoA manager after a successful execution of
// ConvertSubnetToL1Tx on P-Chain
// needs the list of validators for that tx,
// [convertSubnetValidators], together with an evm [ownerAddress]
// to set as the owner of the PoA manager
func (c *Subnet) InitializeProofOfAuthority(
	log logging.Logger,
	signer *evm.Signer,
	aggregatorLogger logging.Logger,
	signatureAggregatorEndpoint string,
) error {
	if c.Network == network.UndefinedNetwork {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingNetwork)
	}
	if c.SubnetID == ids.Empty {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingSubnetID)
	}
	if c.ValidatorManagerBlockchainID == ids.Empty {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingValidatorManagerBlockchainID)
	}
	if c.ValidatorManagerRPC == "" {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingValidatorManagerRPC)
	}
	if c.ValidatorManagerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingValidatorManagerAddress)
	}
	if c.ValidatorManagerOwnerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingValidatorManagerOwnerAddress)
	}

	if len(c.BootstrapValidators) == 0 {
		return fmt.Errorf("unable to initialize Proof of Authority: %w", errMissingBootstrapValidators)
	}

	if client, err := evm.GetClient(c.ValidatorManagerRPC); err != nil {
		log.Error("failure connecting to Validator Manager RPC to setup proposer VM", zap.Error(err))
	} else {
		if err := client.SetupProposerVM(signer); err != nil {
			log.Error("failure setting proposer VM on Validator Manager's Blockchain", zap.Error(err))
		}
		client.Close()
	}

	tx, _, err := validatormanager.PoAValidatorManagerInitialize(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		signer,
		c.SubnetID,
		*c.ValidatorManagerOwnerAddress,
	)
	if err != nil {
		if !errors.Is(err, validatormanager.ErrAlreadyInitialized) {
			return evm.TransactionError(tx, err, "failure initializing poa validator manager")
		}
		log.Info("the PoA contract is already initialized, skipping initializing Proof of Authority contract")
	}

	subnetConversionUnsignedMessage, err := validatormanager.GetPChainSubnetToL1ConversionUnsignedMessage(
		c.Network,
		c.SubnetID,
		c.ValidatorManagerBlockchainID,
		*c.ValidatorManagerAddress,
		c.BootstrapValidators,
	)
	if err != nil {
		return fmt.Errorf("failure signing subnet conversion warp message: %w", err)
	}

	messageHexStr := hex.EncodeToString(subnetConversionUnsignedMessage.Bytes())
	justificationHexStr := hex.EncodeToString(c.SubnetID[:])

	l1Epoch, err := utils.GetCurrentL1Epoch(c.ValidatorManagerRPC, c.ValidatorManagerBlockchainID.String())
	if err != nil {
		return fmt.Errorf("failure getting p-chain height: %w", err)
	}

	signedMessage, err := interchain.SignMessage(
		aggregatorLogger,
		signatureAggregatorEndpoint,
		messageHexStr,
		justificationHexStr,
		c.ValidatorManagerSubnetID.String(),
		0,
		l1Epoch.PChainHeight,
	)
	if err != nil {
		return fmt.Errorf("failed to get signed message: %w", err)
	}
	tx, _, err = validatormanager.InitializeValidatorsSet(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		signer,
		c.SubnetID,
		c.ValidatorManagerBlockchainID,
		c.BootstrapValidators,
		signedMessage,
	)
	if err != nil {
		return evm.TransactionError(tx, err, "failure initializing validators set on poa manager")
	}

	return nil
}

func (c *Subnet) InitializeProofOfStakeNative(
	log logging.Logger,
	signer *evm.Signer,
	aggregatorLogger logging.Logger,
	posParams validatormanager.PoSParams,
	signatureAggregatorEndpoint string,
	nativeMinterPrecompileAdminSigner *evm.Signer,
) error {
	if c.Network == network.UndefinedNetwork {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingNetwork)
	}
	if c.SubnetID == ids.Empty {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingSubnetID)
	}
	if c.ValidatorManagerBlockchainID == ids.Empty {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingValidatorManagerBlockchainID)
	}
	if c.ValidatorManagerRPC == "" {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingValidatorManagerRPC)
	}
	if c.SpecializedValidatorManagerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingSpecializedValidatorManagerAddress)
	}
	if c.ValidatorManagerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingValidatorManagerAddress)
	}
	if c.ValidatorManagerOwnerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingValidatorManagerOwnerAddress)
	}
	if c.ValidatorManagerOwnerSigner == nil {
		return fmt.Errorf("unable to initialize Proof of Stake: %w", errMissingValidatorManagerOwnerPrivateKey)
	}
	if client, err := evm.GetClient(c.ValidatorManagerRPC); err != nil {
		log.Error("failure connecting to Validator Manager RPC to setup proposer VM", zap.Error(err))
	} else {
		if err := client.SetupProposerVM(signer); err != nil {
			log.Error("failure setting proposer VM on Validator Manager's Blockchain", zap.Error(err))
		}
		client.Close()
	}
	tx, _, err := validatormanager.PoAValidatorManagerInitialize(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		signer,
		c.SubnetID,
		*c.ValidatorManagerOwnerAddress,
	)
	if err != nil {
		if !errors.Is(err, validatormanager.ErrAlreadyInitialized) {
			return evm.TransactionError(tx, err, "failure initializing validator manager")
		}
		log.Info("the Validator Manager contract is already initialized, skipping initializing it")
	}
	tx, _, err = validatormanager.PoSValidatorManagerInitialize(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		*c.SpecializedValidatorManagerAddress,
		c.ValidatorManagerOwnerSigner,
		signer,
		posParams,
		nativeMinterPrecompileAdminSigner,
	)
	if err != nil {
		if !errors.Is(err, validatormanager.ErrAlreadyInitialized) {
			return evm.TransactionError(tx, err, "failure initializing native PoS validator manager")
		}
		log.Info("the PoS contract is already initialized, skipping initializing Proof of Stake contract")
	}
	subnetConversionUnsignedMessage, err := validatormanager.GetPChainSubnetToL1ConversionUnsignedMessage(
		c.Network,
		c.SubnetID,
		c.ValidatorManagerBlockchainID,
		*c.ValidatorManagerAddress,
		c.BootstrapValidators,
	)
	if err != nil {
		return fmt.Errorf("failure signing subnet conversion warp message: %w", err)
	}

	messageHexStr := hex.EncodeToString(subnetConversionUnsignedMessage.Bytes())
	justificationHexStr := hex.EncodeToString(c.SubnetID[:])

	l1Epoch, err := utils.GetCurrentL1Epoch(c.ValidatorManagerRPC, c.ValidatorManagerBlockchainID.String())
	if err != nil {
		return fmt.Errorf("failure getting p-chain height: %w", err)
	}
	signedMessage, err := interchain.SignMessage(
		aggregatorLogger,
		signatureAggregatorEndpoint,
		messageHexStr,
		justificationHexStr,
		c.ValidatorManagerSubnetID.String(),
		0,
		l1Epoch.PChainHeight,
	)
	if err != nil {
		return fmt.Errorf("failed to get signed message: %w", err)
	}

	tx, _, err = validatormanager.InitializeValidatorsSet(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		signer,
		c.SubnetID,
		c.ValidatorManagerBlockchainID,
		c.BootstrapValidators,
		signedMessage,
	)
	if err != nil {
		return evm.TransactionError(tx, err, "failure initializing validators set on pos manager")
	}
	return nil
}

func (c *Subnet) InitializeProofOfStakeERC20(
	log logging.Logger,
	signer *evm.Signer,
	aggregatorLogger logging.Logger,
	posParams validatormanager.PoSParams,
	erc20TokenAddress common.Address,
	signatureAggregatorEndpoint string,
) error {
	if c.Network == network.UndefinedNetwork {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingNetwork)
	}
	if c.SubnetID == ids.Empty {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingSubnetID)
	}
	if c.ValidatorManagerBlockchainID == ids.Empty {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingValidatorManagerBlockchainID)
	}
	if c.ValidatorManagerRPC == "" {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingValidatorManagerRPC)
	}
	if c.SpecializedValidatorManagerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingSpecializedValidatorManagerAddress)
	}
	if c.ValidatorManagerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingValidatorManagerAddress)
	}
	if c.ValidatorManagerOwnerAddress == nil {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingValidatorManagerOwnerAddress)
	}
	if c.ValidatorManagerOwnerSigner == nil {
		return fmt.Errorf("unable to initialize Proof of Stake ERC20: %w", errMissingValidatorManagerOwnerPrivateKey)
	}
	if client, err := evm.GetClient(c.ValidatorManagerRPC); err != nil {
		log.Error("failure connecting to Validator Manager RPC to setup proposer VM", zap.Error(err))
	} else {
		if err := client.SetupProposerVM(signer); err != nil {
			log.Error("failure setting proposer VM on Validator Manager's Blockchain", zap.Error(err))
		}
		client.Close()
	}
	tx, _, err := validatormanager.PoAValidatorManagerInitialize(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		signer,
		c.SubnetID,
		*c.ValidatorManagerOwnerAddress,
	)
	if err != nil {
		if !errors.Is(err, validatormanager.ErrAlreadyInitialized) {
			return evm.TransactionError(tx, err, "failure initializing validator manager")
		}
		log.Info("the Validator Manager contract is already initialized, skipping initializing it")
	}
	if err := validatormanager.PoSERC20ValidatorManagerInitialize(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		*c.SpecializedValidatorManagerAddress,
		c.ValidatorManagerOwnerSigner,
		signer,
		posParams,
		erc20TokenAddress,
	); err != nil {
		if !errors.Is(err, validatormanager.ErrAlreadyInitialized) {
			return fmt.Errorf("failure initializing ERC20 PoS validator manager: %w", err)
		}
		log.Info("the ERC20 PoS contract is already initialized, skipping initializing Proof of Stake contract")
	}
	subnetConversionUnsignedMessage, err := validatormanager.GetPChainSubnetToL1ConversionUnsignedMessage(
		c.Network,
		c.SubnetID,
		c.ValidatorManagerBlockchainID,
		*c.ValidatorManagerAddress,
		c.BootstrapValidators,
	)
	if err != nil {
		return fmt.Errorf("failure signing subnet conversion warp message: %w", err)
	}

	messageHexStr := hex.EncodeToString(subnetConversionUnsignedMessage.Bytes())
	justificationHexStr := hex.EncodeToString(c.SubnetID[:])

	l1Epoch, err := utils.GetCurrentL1Epoch(c.ValidatorManagerRPC, c.ValidatorManagerBlockchainID.String())
	if err != nil {
		return fmt.Errorf("failure getting p-chain height: %w", err)
	}
	signedMessage, err := interchain.SignMessage(
		aggregatorLogger,
		signatureAggregatorEndpoint,
		messageHexStr,
		justificationHexStr,
		c.ValidatorManagerSubnetID.String(),
		0,
		l1Epoch.PChainHeight,
	)
	if err != nil {
		return fmt.Errorf("failed to get signed message: %w", err)
	}

	tx, _, err = validatormanager.InitializeValidatorsSet(
		log,
		c.ValidatorManagerRPC,
		*c.ValidatorManagerAddress,
		signer,
		c.SubnetID,
		c.ValidatorManagerBlockchainID,
		c.BootstrapValidators,
		signedMessage,
	)
	if err != nil {
		return evm.TransactionError(tx, err, "failure initializing validators set on ERC20 pos manager")
	}
	return nil
}
