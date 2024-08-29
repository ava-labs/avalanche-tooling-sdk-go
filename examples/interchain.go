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
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
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
	icmVersion, err := icm.GetLatestVersion()
	if err != nil {
		return err
	}
	td := icm.Deployer{}
	if err := td.DownloadAssets(icmVersion); err != nil {
		return err
	}
	chain1MessengerAlreadyDeployed, chain1MessengerAddress, chain1RegistryAddress, err := td.Deploy(
		chain1RPC,
		chain1PK,
		true,
		true,
	)
	if err != nil {
		return err
	}
	if !chain1MessengerAlreadyDeployed {
		return fmt.Errorf("icm already deployed to %s", chain1RPC)
	}
	chain2MessengerAlreadyDeployed, chain2MessengerAddress, chain2RegistryAddress, err := td.Deploy(
		chain2RPC,
		chain2PK,
		true,
		true,
	)
	if err != nil {
		return err
	}
	if !chain2MessengerAlreadyDeployed {
		return fmt.Errorf("icm already deployed to %s", chain2RPC)
	}

	chain1RegistryAddress = "0x4bC756894C6CB10A5735816E25132486F5a1cE8f"
	chain2RegistryAddress = "0x302a91b43d974Cd6f12f4Eae8cADBc8efB7359c8"

	chain1RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}
	chain2RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}

	relayerStorageDir := filepath.Join(relayerDir, "storage")

	relayerConfig := relayer.CreateBaseRelayerConfig(
		logging.Info.LowerString(),
		relayerStorageDir,
		0,
		network,
	)
	relayer.AddBlockchainToRelayerConfig(
		relayerConfig,
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
		relayerConfig,
		chain2RPC,
		"",
		chain2SubnetID,
		chain2BlockchainID,
		chain2RegistryAddress,
		chain2MessengerAddress,
		chain2RelayerKey.C(),
		chain2RelayerKey.PrivKeyHex(),
	)
	relayerConfigPath := filepath.Join(relayerDir, "config.json")
	if err := relayer.SaveRelayerConfig(relayerConfig, relayerConfigPath); err != nil {
		return err
	}

	desiredRelayerBalance := big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10)) // 10 TOKENS

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

	binPath, err := relayer.InstallLatest(relayerDir, "")
	if err != nil {
		return err
	}

	relayerLogPath := filepath.Join(relayerDir, "log.json")

	pid, err := relayer.Execute(binPath, relayerConfigPath, relayerLogPath, "")
	if err != nil {
		if bs, err := os.ReadFile(relayerLogPath); err == nil {
			fmt.Println(string(bs))
		}
		return err
	}

	if err := relayer.WaitForInitialization(relayerConfigPath, relayerLogPath, 0, 0); err != nil {
		return err
	}

	message := "hello world"
	encodedMessage := []byte("hello world")
	tx, receipt, err := icm.SendCrossChainMessage(
		chain1RPC,
		common.HexToAddress(chain1MessengerAddress),
		chain1PK,
		chain2BlockchainID,
		common.Address{},
		encodedMessage,
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

	event, err := evm.GetEventFromLogs(receipt.Logs, icm.ParseSendCrossChainMessage)
	if err != nil {
		return err
	}

	if chain2BlockchainID != ids.ID(event.DestinationBlockchainID[:]) {
		return fmt.Errorf("invalid destination blockchain id at source event, expected %s, got %s", chain2BlockchainID, ids.ID(event.DestinationBlockchainID[:]))
	}
	if message != string(event.Message.Message) {
		return fmt.Errorf("invalid message content at source event, expected %s, got %s", message, string(event.Message.Message))
	}

	if err := icm.WaitForMessageReception(
		chain2RPC,
		chain2MessengerAddress,
		event.MessageID,
		0,
		0,
	); err != nil {
		return err
	}

	return relayer.Cleanup(pid, "", relayerStorageDir)
}
