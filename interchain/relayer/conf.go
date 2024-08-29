// Copyright (C) 2022, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.
package relayer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/awm-relayer/config"
	offchainregistry "github.com/ava-labs/awm-relayer/messages/off-chain-registry"
	"github.com/ethereum/go-ethereum/crypto"
)

var defaultRequiredBalance = big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(500)) // 500 AVAX

func GetSourceConfig(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
) *config.SourceBlockchain {
	rpcEndpoint := network.BlockchainEndpoint(blockchainID.String())
	p := utils.Find(
		relayerConfig.SourceBlockchains,
		func(s *config.SourceBlockchain) bool {
			return s.BlockchainID == blockchainID.String() && s.SubnetID == subnetID.String() && s.RPCEndpoint.BaseURL == rpcEndpoint
		},
	)
	if p != nil {
		return *p
	}
	return nil
}

func GetDestinationConfig(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
) *config.DestinationBlockchain {
	rpcEndpoint := network.BlockchainEndpoint(blockchainID.String())
	p := utils.Find(
		relayerConfig.DestinationBlockchains,
		func(s *config.DestinationBlockchain) bool {
			return s.BlockchainID == blockchainID.String() && s.SubnetID == subnetID.String() && s.RPCEndpoint.BaseURL == rpcEndpoint
		},
	)
	if p != nil {
		return *p
	}
	return nil
}

// fund the relayer private key associated to [network]/[subnetID]/[blockchainID]
// at [relayerConfig]. if [amount] > 0 transfers it to the account. if,
// afterwards, balance < [requiredMinBalance], transfers remaining amount for that
// if [requiredMinBalance] is nil, uses [defaultRequiredBalance]
func FundRelayer(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	privateKey string,
	amount *big.Int,
	requiredMinBalance *big.Int,
) error {
	destinationConfig := GetDestinationConfig(relayerConfig, network, subnetID, blockchainID)
	if destinationConfig == nil {
		return fmt.Errorf("relayer destination not found for network %s, subnet %s, blockchain %s", network.Kind.String(), subnetID.String(), blockchainID.String())
	}
	return FundRelayerPrivateKey(
		network,
		blockchainID,
		privateKey,
		destinationConfig.AccountPrivateKey,
		amount,
		requiredMinBalance,
	)
}

// fund [relayerPrivateKey] at [network]/[blockchainID]
// see FundRelayer for [amount]/[requiredMinBalance] logic
func FundRelayerPrivateKey(
	network avalanche.Network,
	blockchainID ids.ID,
	privateKey string,
	relayerPrivateKey string,
	amount *big.Int,
	requiredMinBalance *big.Int,
) error {
	relayerPK, err := crypto.HexToECDSA(relayerPrivateKey)
	if err != nil {
		return err
	}
	relayerAddress := crypto.PubkeyToAddress(relayerPK.PublicKey)
	return FundRelayerAddress(
		network,
		blockchainID,
		privateKey,
		relayerAddress.Hex(),
		amount,
		requiredMinBalance,
	)
}

// fund [relayerAddress] at [network]/[blockchainID]
// see FundRelayer for [amount]/[requiredMinBalance] logic
func FundRelayerAddress(
	network avalanche.Network,
	blockchainID ids.ID,
	privateKey string,
	relayerAddress string,
	amount *big.Int,
	requiredMinBalance *big.Int,
) error {
	if requiredMinBalance == nil {
		requiredMinBalance = defaultRequiredBalance
	}
	client, err := evm.GetClient(network.BlockchainEndpoint(blockchainID.String()))
	if err != nil {
		return err
	}
	if amount != nil && amount.Cmp(big.NewInt(0)) > 0 {
		if err := evm.Transfer(
			client,
			privateKey,
			relayerAddress,
			amount,
		); err != nil {
			return err
		}
	}
	return evm.SetMinBalance(
		client,
		privateKey,
		relayerAddress,
		requiredMinBalance,
	)
}

func LoadRelayerConfig(relayerConfigPath string) (*config.Config, error) {
	relayerConfig := config.Config{}
	bs, err := os.ReadFile(relayerConfigPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &relayerConfig); err != nil {
		return nil, err
	}
	return &relayerConfig, nil
}

func SaveRelayerConfig(relayerConfig *config.Config, relayerConfigPath string) error {
	bs, err := json.MarshalIndent(relayerConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(relayerConfigPath, bs, constants.WriteReadReadPerms)
}

func CreateBaseRelayerConfigFile(
	relayerConfigPath string,
	logLevel string,
	storageLocation string,
	metricsPort uint16,
	network avalanche.Network,
) error {
	relayerConfig := CreateBaseRelayerConfig(
		logLevel,
		storageLocation,
		metricsPort,
		network,
	)
	return SaveRelayerConfig(relayerConfig, relayerConfigPath)
}

func AddSourceToRelayerConfigFile(
	relayerConfigPath string,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	icmRegistryAddress string,
	icmMessengerAddress string,
	relayerRewardAddress string,
) error {
	relayerConfig, err := LoadRelayerConfig(relayerConfigPath)
	if err != nil {
		return err
	}
	AddSourceToRelayerConfig(
		relayerConfig,
		network,
		subnetID,
		blockchainID,
		icmRegistryAddress,
		icmMessengerAddress,
		relayerRewardAddress,
	)
	return SaveRelayerConfig(relayerConfig, relayerConfigPath)
}

func AddDestinationToRelayerConfigFile(
	relayerConfigPath string,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	relayerPrivateKey string,
) error {
	relayerConfig, err := LoadRelayerConfig(relayerConfigPath)
	if err != nil {
		return err
	}
	AddDestinationToRelayerConfig(
		relayerConfig,
		network,
		subnetID,
		blockchainID,
		relayerPrivateKey,
	)
	return SaveRelayerConfig(relayerConfig, relayerConfigPath)
}

func AddBlockchainToRelayerConfigFile(
	relayerConfigPath string,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	icmRegistryAddress string,
	icmMessengerAddress string,
	relayerRewardAddress string,
	relayerPrivateKey string,
) error {
	relayerConfig, err := LoadRelayerConfig(relayerConfigPath)
	if err != nil {
		return err
	}
	AddBlockchainToRelayerConfig(
		relayerConfig,
		network,
		subnetID,
		blockchainID,
		icmRegistryAddress,
		icmMessengerAddress,
		relayerRewardAddress,
		relayerPrivateKey,
	)
	return SaveRelayerConfig(relayerConfig, relayerConfigPath)
}

func CreateBaseRelayerConfig(
	logLevel string,
	storageLocation string,
	metricsPort uint16,
	network avalanche.Network,
) *config.Config {
	relayerConfig := &config.Config{
		LogLevel: logLevel,
		PChainAPI: &config.APIConfig{
			BaseURL:     network.Endpoint,
			QueryParams: map[string]string{},
		},
		InfoAPI: &config.APIConfig{
			BaseURL:     network.Endpoint,
			QueryParams: map[string]string{},
		},
		StorageLocation:        storageLocation,
		ProcessMissedBlocks:    false,
		SourceBlockchains:      []*config.SourceBlockchain{},
		DestinationBlockchains: []*config.DestinationBlockchain{},
	}
	if metricsPort != 0 {
		relayerConfig.MetricsPort = metricsPort
	}
	return relayerConfig
}

func AddSourceToRelayerConfig(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	icmRegistryAddress string,
	icmMessengerAddress string,
	relayerRewardAddress string,
) {
	source := &config.SourceBlockchain{
		SubnetID:     subnetID.String(),
		BlockchainID: blockchainID.String(),
		VM:           config.EVM.String(),
		RPCEndpoint: config.APIConfig{
			BaseURL: network.BlockchainEndpoint(blockchainID.String()),
		},
		WSEndpoint: config.APIConfig{
			BaseURL: network.BlockchainWSEndpoint(blockchainID.String()),
		},
		MessageContracts: map[string]config.MessageProtocolConfig{
			icmMessengerAddress: {
				MessageFormat: config.TELEPORTER.String(),
				Settings: map[string]interface{}{
					"reward-address": relayerRewardAddress,
				},
			},
			offchainregistry.OffChainRegistrySourceAddress.Hex(): {
				MessageFormat: config.OFF_CHAIN_REGISTRY.String(),
				Settings: map[string]interface{}{
					"teleporter-registry-address": icmRegistryAddress,
				},
			},
		},
	}
	if GetSourceConfig(relayerConfig, network, subnetID, blockchainID) == nil {
		relayerConfig.SourceBlockchains = append(relayerConfig.SourceBlockchains, source)
	}
}

func AddDestinationToRelayerConfig(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	relayerFundedAddressKey string,
) {
	destination := &config.DestinationBlockchain{
		SubnetID:     subnetID.String(),
		BlockchainID: blockchainID.String(),
		VM:           config.EVM.String(),
		RPCEndpoint: config.APIConfig{
			BaseURL: network.BlockchainEndpoint(blockchainID.String()),
		},
		AccountPrivateKey: relayerFundedAddressKey,
	}
	if GetDestinationConfig(relayerConfig, network, subnetID, blockchainID) == nil {
		relayerConfig.DestinationBlockchains = append(relayerConfig.DestinationBlockchains, destination)
	}
}

func AddBlockchainToRelayerConfig(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	icmRegistryAddress string,
	icmMessengerAddress string,
	relayerRewardAddress string,
	relayerPrivateKey string,
) {
	AddSourceToRelayerConfig(
		relayerConfig,
		network,
		subnetID,
		blockchainID,
		icmRegistryAddress,
		icmMessengerAddress,
		relayerRewardAddress,
	)
	AddDestinationToRelayerConfig(
		relayerConfig,
		network,
		subnetID,
		blockchainID,
		relayerPrivateKey,
	)
}
