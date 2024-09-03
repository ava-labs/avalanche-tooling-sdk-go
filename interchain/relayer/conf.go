// Copyright (C) 2022, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.
package relayer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/awm-relayer/config"
	offchainregistry "github.com/ava-labs/awm-relayer/messages/off-chain-registry"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetSourceConfig(
	relayerConfig *config.Config,
	blockchainID ids.ID,
) *config.SourceBlockchain {
	p := utils.Find(
		relayerConfig.SourceBlockchains,
		func(s *config.SourceBlockchain) bool {
			return s.BlockchainID == blockchainID.String()
		},
	)
	if p != nil {
		return *p
	}
	return nil
}

func GetDestinationConfig(
	relayerConfig *config.Config,
	blockchainID ids.ID,
) *config.DestinationBlockchain {
	p := utils.Find(
		relayerConfig.DestinationBlockchains,
		func(s *config.DestinationBlockchain) bool {
			return s.BlockchainID == blockchainID.String()
		},
	)
	if p != nil {
		return *p
	}
	return nil
}

// FundRelayer funds the relayer private key associated to [blockchainID] specified in
// [relayerConfig]. Receives one of two amount specs:
// 1) if [amount] > 0, transfers it to the account.
// 2) if [requiredMinBalance] > 0, checks the balance in the account, and
// if balance < [requiredMinBalance], transfers the amount needed so as
// balance == [requiredMinBalance]
func FundRelayer(
	relayerConfig *config.Config,
	blockchainID ids.ID,
	privateKey string,
	amount *big.Int,
	requiredMinBalance *big.Int,
) error {
	destinationConfig := GetDestinationConfig(relayerConfig, blockchainID)
	if destinationConfig == nil {
		return fmt.Errorf("relayer destination not found for blockchain %s", blockchainID.String())
	}
	return FundRelayerPrivateKey(
		destinationConfig.RPCEndpoint.BaseURL,
		privateKey,
		destinationConfig.AccountPrivateKey,
		amount,
		requiredMinBalance,
	)
}

// FundRelayerPrivateKey funds [relayerPrivateKey] at [rpcEndpoint]
// see FundRelayer for [amount]/[requiredMinBalance] logic
func FundRelayerPrivateKey(
	rpcEndpoint string,
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
		rpcEndpoint,
		privateKey,
		relayerAddress.Hex(),
		amount,
		requiredMinBalance,
	)
}

// FundRelayerAddress funds [relayerAddress] at [rpcEndpoint]
// see FundRelayer for [amount]/[requiredMinBalance] logic
func FundRelayerAddress(
	rpcEndpoint string,
	privateKey string,
	relayerAddress string,
	amount *big.Int,
	requiredMinBalance *big.Int,
) error {
	if amount == nil && requiredMinBalance == nil {
		return fmt.Errorf("failure funding relayer: you must provide an amount or a required balance")
	}
	client, err := evm.GetClient(rpcEndpoint)
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

func UnserializeRelayerConfig(relayerConfigBytes []byte) (*config.Config, error) {
	relayerConfig := config.Config{}
	if err := json.Unmarshal(relayerConfigBytes, &relayerConfig); err != nil {
		return nil, err
	}
	return &relayerConfig, nil
}

func SerializeRelayerConfig(relayerConfig *config.Config) ([]byte, error) {
	return json.MarshalIndent(relayerConfig, "", "  ")
}

// Creates a base relayer config
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
		MetricsPort:            metricsPort,
		DBWriteIntervalSeconds: 1,
	}
	return relayerConfig
}

// Adds  blockchain to the relayer config, setting it as source.
// So the relayer will listed for new messages in it,
// sending those to other blockchains.
func AddSourceToRelayerConfig(
	relayerConfig *config.Config,
	rpcEndpoint string,
	wsEndpoint string,
	subnetID ids.ID,
	blockchainID ids.ID,
	icmRegistryAddress string,
	icmMessengerAddress string,
	relayerRewardAddress string,
) {
	if wsEndpoint == "" {
		wsEndpoint = strings.TrimPrefix(rpcEndpoint, "https")
		wsEndpoint = strings.TrimPrefix(wsEndpoint, "http")
		wsEndpoint = strings.TrimSuffix(wsEndpoint, "rpc")
		wsEndpoint = fmt.Sprintf("%s%s%s", "ws", wsEndpoint, "ws")
	}
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
	if GetSourceConfig(relayerConfig, blockchainID) == nil {
		relayerConfig.SourceBlockchains = append(relayerConfig.SourceBlockchains, source)
	}
}

// Adds a blockchain to the relayer config,
// setting it as destination.
// So the relayer will send to it new messages from other blockchains.
func AddDestinationToRelayerConfig(
	relayerConfig *config.Config,
	rpcEndpoint string,
	subnetID ids.ID,
	blockchainID ids.ID,
	relayerFundedAddressKey string,
) {
	destination := &config.DestinationBlockchain{
		SubnetID:     subnetID.String(),
		BlockchainID: blockchainID.String(),
		VM:           config.EVM.String(),
		RPCEndpoint: config.APIConfig{
			BaseURL: rpcEndpoint,
		},
		AccountPrivateKey: relayerFundedAddressKey,
	}
	if GetDestinationConfig(relayerConfig, blockchainID) == nil {
		relayerConfig.DestinationBlockchains = append(relayerConfig.DestinationBlockchains, destination)
	}
}

// Adds a blockchain to the relayer config,
// setting it both as source and as destination.
// So the relayer will both listed for new messages in it,
// and send to it new messages from other blockchains.
func AddBlockchainToRelayerConfig(
	relayerConfig *config.Config,
	rpcEndpoint string,
	wsEndpoint string,
	subnetID ids.ID,
	blockchainID ids.ID,
	icmRegistryAddress string,
	icmMessengerAddress string,
	relayerRewardAddress string,
	relayerPrivateKey string,
) {
	AddSourceToRelayerConfig(
		relayerConfig,
		rpcEndpoint,
		wsEndpoint,
		subnetID,
		blockchainID,
		icmRegistryAddress,
		icmMessengerAddress,
		relayerRewardAddress,
	)
	AddDestinationToRelayerConfig(
		relayerConfig,
		rpcEndpoint,
		subnetID,
		blockchainID,
		relayerPrivateKey,
	)
}
