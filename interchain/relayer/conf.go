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

func GetBaseRelayerConfig(
	logLevel string,
	storageLocation string,
	network avalanche.Network,
) *config.Config {
	return &config.Config{
		LogLevel: logLevel,
		PChainAPI: &config.APIConfig{
			BaseURL: network.Endpoint,
		},
		InfoAPI: &config.APIConfig{
			BaseURL: network.Endpoint,
		},
		StorageLocation:        storageLocation,
		ProcessMissedBlocks:    false,
		SourceBlockchains:      []*config.SourceBlockchain{},
		DestinationBlockchains: []*config.DestinationBlockchain{},
	}
}

func AddChainToRelayerConfig(
	relayerConfig *config.Config,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	teleporterContractAddress string,
	teleporterRegistryAddress string,
	privateKey string,
	rewardAddress string,
) error {
	rpcEndpoint := network.BlockchainEndpoint(blockchainID.String())
	wsEndpoint := network.BlockchainWSEndpoint(blockchainID.String())
	source := &config.SourceBlockchain{
		SubnetID:     subnetID.String(),
		BlockchainID: blockchainID.String(),
		VM:           config.EVM.String(),
		RPCEndpoint: config.APIConfig{
			BaseURL: rpcEndpoint,
		},
		WSEndpoint: config.APIConfig{
			BaseURL: wsEndpoint,
		},
		MessageContracts: map[string]config.MessageProtocolConfig{
			teleporterContractAddress: {
				MessageFormat: config.TELEPORTER.String(),
				Settings: map[string]interface{}{
					"reward-address": rewardAddress,
				},
			},
			offchainregistry.OffChainRegistrySourceAddress.Hex(): {
				MessageFormat: config.OFF_CHAIN_REGISTRY.String(),
				Settings: map[string]interface{}{
					"teleporter-registry-address": teleporterRegistryAddress,
				},
			},
		},
	}
	destination := &config.DestinationBlockchain{
		SubnetID:     subnetID.String(),
		BlockchainID: blockchainID.String(),
		VM:           config.EVM.String(),
		RPCEndpoint: config.APIConfig{
			BaseURL: rpcEndpoint,
		},
		AccountPrivateKey: privateKey,
	}
	if GetSourceConfig(relayerConfig, network, subnetID, blockchainID) == nil {
		relayerConfig.SourceBlockchains = append(relayerConfig.SourceBlockchains, source)
	}
	if GetDestinationConfig(relayerConfig, network, subnetID, blockchainID) == nil {
		relayerConfig.DestinationBlockchains = append(relayerConfig.DestinationBlockchains, destination)
	}
	return nil
}

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

// add a blockchain both as source and as destination into the file at [relayerConfigPath]
func AddChainToRelayerConfigFile(
	logLevel string,
	relayerConfigPath string,
	relayerStorageDir string,
	network avalanche.Network,
	subnetID ids.ID,
	blockchainID ids.ID,
	teleporterContractAddress string,
	teleporterRegistryAddress string,
	privateKey string,
	rewardAddress string,
) error {
	awmRelayerConfig := &config.Config{}
	if utils.FileExists(relayerConfigPath) {
		bs, err := os.ReadFile(relayerConfigPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bs, &awmRelayerConfig); err != nil {
			return err
		}
	} else {
		awmRelayerConfig = GetBaseRelayerConfig(
			logLevel,
			relayerStorageDir,
			network,
		)
	}
	if err := AddChainToRelayerConfig(
		awmRelayerConfig,
		network,
		subnetID,
		blockchainID,
		teleporterContractAddress,
		teleporterRegistryAddress,
		privateKey,
		rewardAddress,
	); err != nil {
		return err
	}
	bs, err := json.MarshalIndent(awmRelayerConfig, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(relayerConfigPath, bs, constants.WriteReadReadPerms); err != nil {
		return err
	}
	return nil
}
