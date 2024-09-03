// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/interchainmessenger"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer/localrelayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/awm-relayer/config"
	"github.com/ethereum/go-ethereum/common"
)

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
	// Deploy Interchain Messenger (ICM)
	// More information: https://github.com/ava-labs/teleporter
	fmt.Println("Deploying Interchain Messenger")
	chain1RegistryAddress, chain1MessengerAddress, chain2RegistryAddress, chain2MessengerAddress, err := SetupICM(
		chain1RPC,
		chain1PK,
		chain2RPC,
		chain2PK,
	)
	if err != nil {
		return err
	}

	// Create keys for relayer operations: pay fees and receive rewards
	// At destination blockchain, the relayer needs to pay fees for smart contract
	// calls used to deliver messages. For that, it needs information on properly
	// funded private keys at destination.
	// Besides this, some source blockchain may provide incentives to relayers
	// that send their messages. For that, the relayer needs to be configured with
	// the address that will receive such payments at source.
	// More information: https: //github.com/ava-labs/awm-relayer/blob/main/relayer/README.md
	chain1RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}
	chain2RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}

	// Creates relayer config for the two chains
	// The relayer can be configured to listed for new ICM messages
	// from a set of source blockchains, and then deliver those
	// to a set of destination blockchains.
	// Here we are configuring chain1 and chain2 both as source
	// and as destination, so we can send messages in any direction.
	relayerConfigPath := filepath.Join(relayerDir, "config.json")
	relayerStorageDir := filepath.Join(relayerDir, "storage")
	relayerConfig, relayerBytes, err := SetupRelayerConf(
		relayerStorageDir,
		network,
		chain1RPC,
		chain1SubnetID,
		chain1BlockchainID,
		chain1RegistryAddress,
		chain1MessengerAddress,
		chain1RelayerKey,
		chain2RPC,
		chain2SubnetID,
		chain2BlockchainID,
		chain2RegistryAddress,
		chain2MessengerAddress,
		chain2RelayerKey,
	)
	if err != nil {
		return err
	}
	if err := os.WriteFile(relayerConfigPath, relayerBytes, constants.WriteReadReadPerms); err != nil {
		return err
	}
	fmt.Printf("Generated relayer conf on %s\n", relayerConfigPath)

	// Fund each relayer key with 10 TOKENs
	// Where TOKEN is the native gas token of each blockchain
	// Assumes that the TOKEN decimals are 18, so, this equals
	// to 1e18 of the smallest gas amount in each chain
	fmt.Printf("Funding relayer keys %s, %s\n", chain1RelayerKey.C(), chain2RelayerKey.C())
	desiredRelayerBalance := big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10))
	// chain1PK will have a balance 10 native gas tokens on chain.
	if err := relayer.FundRelayer(
		relayerConfig,
		chain1BlockchainID,
		chain1PK,
		nil,
		desiredRelayerBalance,
	); err != nil {
		return err
	}
	// chain2PK will have a balance 10 native gas tokens on chain2
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
	defer func() { _ = localrelayer.Cleanup(pid, "", relayerStorageDir) }()

	// send a message from chain1 to chain2
	fmt.Println("Verifying message delivery")
	if err := TestMessageDelivery(
		chain1RPC,
		chain1PK,
		chain1MessengerAddress,
		chain2BlockchainID,
		chain2RPC,
		chain2MessengerAddress,
		[]byte("hello world"),
	); err != nil {
		return err
	}

	fmt.Println("Message successfully delivered")

	return nil
}

func SetupICM(
	chain1RPC string,
	chain1PK string,
	chain2RPC string,
	chain2PK string,
) (string, string, string, string, error) {
	// Get latest version of ICM
	icmVersion, err := interchainmessenger.GetLatestVersion()
	if err != nil {
		return "", "", "", "", err
	}
	// Deploys ICM Messenger and Registry to Chain1 and Chain2
	td := interchainmessenger.Deployer{}
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
	storageDir string,
	network avalanche.Network,
	chain1RPC string,
	chain1SubnetID ids.ID,
	chain1BlockchainID ids.ID,
	chain1RegistryAddress string,
	chain1MessengerAddress string,
	chain1RelayerKey *key.SoftKey,
	chain2RPC string,
	chain2SubnetID ids.ID,
	chain2BlockchainID ids.ID,
	chain2RegistryAddress string,
	chain2MessengerAddress string,
	chain2RelayerKey *key.SoftKey,
) (*config.Config, []byte, error) {
	// Create a base relayer config
	config := relayer.CreateBaseRelayerConfig(
		logging.Info.LowerString(),
		storageDir,
		0,
		network,
	)
	// Add blockchain chain1 to the relayer config,
	// setting it both as source and as destination.
	// So the relayer will both listed for new messages in it,
	// and send to it new messages from other blockchains.
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
	// Add blockchain chain2 to the relayer config,
	// setting it both as source and as destination.
	// So the relayer will both listed for new messages in it,
	// and send to it new messages from other blockchains.
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
	bs, err := relayer.SerializeRelayerConfig(config)
	return config, bs, err
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
	tx, receipt, err := interchainmessenger.SendCrossChainMessage(
		chain1RPC,
		common.HexToAddress(chain1MessengerAddress),
		chain1PK,
		chain2BlockchainID,
		common.Address{},
		message,
	)
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
	if err != nil {
		return err
	}

	// get from chain1 event logs the message id
	event, err := evm.GetEventFromLogs(receipt.Logs, interchainmessenger.ParseSendCrossChainMessage)
	if err != nil {
		return err
	}
	messageID := event.MessageID
	fmt.Println("Source Event Destination Blockchain ID: ", ids.ID(event.DestinationBlockchainID[:]))
	fmt.Println("Source Event Message: ", string(event.Message.Message))

	// wait for chain2 to receive the message
	return interchainmessenger.WaitForMessageReception(
		chain2RPC,
		chain2MessengerAddress,
		messageID,
		0,
		0,
	)
}

// Fuji ICM Example
//
// Deploys ICM into CHAIN1_RPC and CHAIN2_RPC,
// paying deploy fees with CHAIN1_PK and CHAIN2_PK
//
// Downloads and executes a relayer in a local process
// and sets it to listen to CHAIN1 and CHAIN2.
// Subnet IDs and Blockchain IDs are provided to fulfill
// relayer conf
//
// All relayer data is saved into an existing RELAYER_DIR
//
// Example environment setup values:
// export CHAIN1_SUBNET_ID=2AgDoogVySMLkkqfMgzxWSKZWVVmXvNRVixQ5WL7fSUC7sRhTH
// export CHAIN1_BLOCKCHAIN_ID=2Z28jccJqQGCPdF5ee8P3aJq2fsrvyy6F4nhQphGJstTjk9ZsR
// export CHAIN1_RPC=http://36.172.121.333:9650/ext/bc/${CHAIN1_BLOCKCHAIN_ID}/rpc
// export CHAIN1_PK=(64 digit hexadecimal)
// export CHAIN2_SUBNET_ID=2YRENAJfNtBPYB6D4xC1pPV8tn2srdytrsJa6JC3neF2Au5FBc
// export CHAIN2_BLOCKCHAIN_ID=NJT2wqkbNNx9T2cAFsqLNfVVZDamyJyrbRS5fxtevBuMLHwJ8
// export CHAIN2_RPC=http://36.172.121.123:9650/ext/bc/${CHAIN2_BLOCKCHAIN_ID}/rpc
// export CHAIN2_PK=3c82b8787e887a8798f922d95a948bcffa8d1989898a9898ffffee1000ed7c21
// export CHAIN2_PK=(64 digit hexadecimal)
// export RELAYER_DIR=~/relayer_rundir/
func Interchain() error {
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
