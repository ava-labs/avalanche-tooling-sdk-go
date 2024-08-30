// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/icm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer/localrelayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/awm-relayer/config"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	if err := CallInterchainExample(); err != nil {
		panic(err)
	}
}

// Fuji ICM Example
//
// Deploys ICM into CHAIN1_RPC and CHAIN2_RPC,
// paying deploy fees with CHAIN1_PK and CHAIN2_PK
//
// Downloads and executes a relayer in a local process
// and sets it to listen to CHAIN1 and CHAIN2.
// Subnet IDs and Blockchain IDs are provided to fullfill
// relayer conf
//
// All relayer data is saved into RELAYER_DIR, that must
// exist beforehand
func CallInterchainExample() error {
	chain1RPC := os.Getenv("CHAIN1_RPC")
	chain1PK := os.Getenv("CHAIN1_PK")
	chain1SubnetID, err := ids.FromString(os.Getenv("CHAIN1_SUBNET_ID"))
	if err != nil {
		return err
	}
	chain1BlockchainID, err := ids.FromString(os.Getenv("CHAIN1_BLOCKCHAIN_ID"))
	if err != nil {
		return err
	}
	chain2RPC := os.Getenv("CHAIN2_RPC")
	chain2PK := os.Getenv("CHAIN2_PK")
	chain2SubnetID, err := ids.FromString(os.Getenv("CHAIN2_SUBNET_ID"))
	if err != nil {
		return err
	}
	chain2BlockchainID, err := ids.FromString(os.Getenv("CHAIN2_BLOCKCHAIN_ID"))
	if err != nil {
		return err
	}
	relayerDir := os.Getenv("RELAYER_DIR")
	if relayerDir == "" {
		return fmt.Errorf("must define RELAYER_DIR env var")
	}
	relayerDir = utils.ExpandHome(relayerDir)
	if !utils.DirectoryExists(relayerDir) {
		return fmt.Errorf("relayer directory %q must exist", relayerDir)
	}
	return InterchainExample(
		avalanche.FujiNetwork(),
		chain1RPC,
		chain1PK,
		chain1SubnetID,
		chain1BlockchainID,
		chain2RPC,
		chain2PK,
		chain2SubnetID,
		chain2BlockchainID,
		relayerDir,
	)
}

// Deploys ICM in two chains
// Deploys a relayes to interconnect them
// Send an example msg
func InterchainExample(
	network avalanche.Network,
	chain1RPC string,
	chain1PK string,
	chain1SubnetID ids.ID,
	chain1BlockchainID ids.ID,
	chain2RPC string,
	chain2PK string,
	chain2SubnetID ids.ID,
	chain2BlockchainID ids.ID,
	relayerDir string,
) error {
	// Deploy ICM
	fmt.Println("Deploying ICM")
	chain1RegistryAddress, chain1MessengerAddress, chain2RegistryAddress, chain2MessengerAddress, err := SetupICM(
		chain1RPC,
		chain1PK,
		chain2RPC,
		chain2PK,
	)
	if err != nil {
		return err
	}

	// Creates a couple of keys for the Relayer
	chain1RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}
	chain2RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}

	// Creates relayer config for the two chains
	relayerConfigPath := filepath.Join(relayerDir, "config.json")
	relayerStorageDir := filepath.Join(relayerDir, "storage")
	relayerConfig, err := SetupRelayerConf(
		relayerConfigPath,
		relayerStorageDir,
		network,
		chain1RPC,
		chain1PK,
		chain1SubnetID,
		chain1BlockchainID,
		chain1RegistryAddress,
		chain1MessengerAddress,
		chain1RelayerKey,
		chain2RPC,
		chain2PK,
		chain2SubnetID,
		chain2BlockchainID,
		chain2RegistryAddress,
		chain2MessengerAddress,
		chain2RelayerKey,
	)
	fmt.Printf("Generated relayer conf on %s\n", relayerConfigPath)

	// Fund the relayer keys with 10 TOKENS each
	fmt.Printf("Funding relayer keys %s, %s\n", chain1RelayerKey.C(), chain2RelayerKey.C())
	desiredRelayerBalance := big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10))
	if err := relayer.FundRelayer(
		relayerConfig,
		chain1BlockchainID,
		chain1PK,
		nil,
		desiredRelayerBalance,
	); err != nil {
		return err
	}
	if err := relayer.FundRelayer(
		relayerConfig,
		chain2BlockchainID,
		chain2PK,
		nil,
		desiredRelayerBalance,
	); err != nil {
		return err
	}

	// install and execute a relayer on localhost
	// also wait for proper initialization
	relayerLogPath := filepath.Join(relayerDir, "logs.txt")
	fmt.Printf("Executing local relayer with logs %s\n", relayerLogPath)
	pid, _, err := StartLocalRelayer(
		relayerConfigPath,
		relayerLogPath,
		relayerDir,
	)
	if err != nil {
		return err
	}

	// defer stopping relayer and cleaning up
	defer localrelayer.Cleanup(pid, "", relayerStorageDir)

	// send a message from chain1 to chain2
	fmt.Println("Verifying message delivery")
	return TestMessageDelivery(
		chain1RPC,
		chain1PK,
		chain1MessengerAddress,
		chain2BlockchainID,
		chain2RPC,
		chain2MessengerAddress,
		[]byte("hello world"),
	)
}

func SetupICM(
	chain1RPC string,
	chain1PK string,
	chain2RPC string,
	chain2PK string,
) (string, string, string, string, error) {
	// Get latest version of ICM
	icmVersion, err := icm.GetLatestVersion()
	if err != nil {
		return "", "", "", "", err
	}
	// Deploys ICM Messenger and Registry to Chain1 and Chain2
	td := icm.Deployer{}
	if err := td.DownloadAssets(icmVersion); err != nil {
		return "", "", "", "", err
	}
	_, chain1RegistryAddress, chain1MessengerAddress, err := td.Deploy(
		chain1RPC,
		chain1PK,
		true,
	)
	if err != nil {
		return "", "", "", "", err
	}
	_, chain2RegistryAddress, chain2MessengerAddress, err := td.Deploy(
		chain2RPC,
		chain2PK,
		true,
	)
	if err != nil {
		return "", "", "", "", err
	}
	return chain1RegistryAddress, chain1MessengerAddress, chain2RegistryAddress, chain2MessengerAddress, nil
}

func SetupRelayerConf(
	configPath string,
	storageDir string,
	network avalanche.Network,
	chain1RPC string,
	chain1PK string,
	chain1SubnetID ids.ID,
	chain1BlockchainID ids.ID,
	chain1RegistryAddress string,
	chain1MessengerAddress string,
	chain1RelayerKey *key.SoftKey,
	chain2RPC string,
	chain2PK string,
	chain2SubnetID ids.ID,
	chain2BlockchainID ids.ID,
	chain2RegistryAddress string,
	chain2MessengerAddress string,
	chain2RelayerKey *key.SoftKey,
) (*config.Config, error) {
	// Creates relayer config
	config := relayer.CreateBaseRelayerConfig(
		logging.Info.LowerString(),
		storageDir,
		0,
		network,
	)
	relayer.AddBlockchainToRelayerConfig(
		config,
		chain1RPC,
		"",
		chain1SubnetID,
		chain1BlockchainID,
		chain1RegistryAddress,
		chain1MessengerAddress,
		chain1RelayerKey.C(),
		chain1RelayerKey.PrivKeyHex(),
	)
	relayer.AddBlockchainToRelayerConfig(
		config,
		chain2RPC,
		"",
		chain2SubnetID,
		chain2BlockchainID,
		chain2RegistryAddress,
		chain2MessengerAddress,
		chain2RelayerKey.C(),
		chain2RelayerKey.PrivKeyHex(),
	)
	if err := relayer.SaveRelayerConfig(config, configPath); err != nil {
		return nil, err
	}
	return config, nil
}

func StartLocalRelayer(
	configPath string,
	logPath string,
	installDir string,
) (int, string, error) {
	binPath, err := localrelayer.InstallLatest(installDir, "")
	if err != nil {
		return 0, "", err
	}
	pid, err := localrelayer.Execute(binPath, configPath, logPath, "")
	if err != nil {
		if bs, err := os.ReadFile(logPath); err == nil {
			fmt.Println(string(bs))
		}
		return pid, binPath, err
	}
	if err := localrelayer.WaitForInitialization(configPath, logPath, 0, 0); err != nil {
		return pid, binPath, err
	}
	return pid, binPath, nil
}

func TestMessageDelivery(
	chain1RPC string,
	chain1PK string,
	chain1MessengerAddress string,
	chain2BlockchainID ids.ID,
	chain2RPC string,
	chain2MessengerAddress string,
	message []byte,
) error {
	// send message request to chain1
	tx, receipt, err := icm.SendCrossChainMessage(
		chain1RPC,
		common.HexToAddress(chain1MessengerAddress),
		chain1PK,
		chain2BlockchainID,
		common.Address{},
		message,
	)
	if err != nil {
		return err
	}
	if err == evm.ErrFailedReceiptStatus {
		txHash := tx.Hash().String()
		trace, err := evm.GetTrace(chain1RPC, txHash)
		if err != nil {
			fmt.Printf("error obtaining tx trace: %s\n", err)
		} else {
			fmt.Printf("trace: %#v\n", trace)
		}
		return fmt.Errorf("source receipt status for tx %s is not ReceiptStatusSuccessful", txHash)
	}

	// get from chain1 event logs the message id
	event, err := evm.GetEventFromLogs(receipt.Logs, icm.ParseSendCrossChainMessage)
	if err != nil {
		return err
	}
	messageID := event.MessageID
	// also veryfies some input params
	if chain2BlockchainID != ids.ID(event.DestinationBlockchainID[:]) {
		return fmt.Errorf("invalid destination blockchain id at source event, expected %s, got %s", chain2BlockchainID, ids.ID(event.DestinationBlockchainID[:]))
	}
	if string(message) != string(event.Message.Message) {
		return fmt.Errorf("invalid message content at source event, expected %s, got %s", message, string(event.Message.Message))
	}

	// wait for chain2 to receive the message
	if err := icm.WaitForMessageReception(
		chain2RPC,
		chain2MessengerAddress,
		messageID,
		0,
		0,
	); err != nil {
		return err
	}
	return nil
}
