// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/icm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
)

func main() {
	icmVersion, err := icm.GetLatestVersion()
	if err != nil {
		panic(err)
	}
	td := icm.Deployer{}
	if err := td.DownloadAssets(icmVersion); err != nil {
		panic(err)
	}
	chain1RPC := os.Getenv("CHAIN1_RPC")
	chain1PK := os.Getenv("CHAIN1_PK")
	chain2RPC := os.Getenv("CHAIN2_RPC")
	chain2PK := os.Getenv("CHAIN2_PK")
	chain1MessengerAlreadyDeployed, chain1MessengerAddress, chain1RegistryAddress, err := td.Deploy(
		chain1RPC,
		chain1PK,
		true,
		true,
	)
	if err != nil {
		panic(err)
	}
	if !chain1MessengerAlreadyDeployed {
		panic(fmt.Errorf("icm already deployed to %s", chain1RPC))
	}
	chain2MessengerAlreadyDeployed, chain2MessengerAddress, chain2RegistryAddress, err := td.Deploy(
		chain2RPC,
		chain2PK,
		true,
		true,
	)
	if err != nil {
		panic(err)
	}
	if !chain2MessengerAlreadyDeployed {
		panic(fmt.Errorf("icm already deployed to %s", chain2RPC))
	}

	chain1RegistryAddress = "0x4bC756894C6CB10A5735816E25132486F5a1cE8f"
	chain2RegistryAddress = "0x302a91b43d974Cd6f12f4Eae8cADBc8efB7359c8"

	network := avalanche.FujiNetwork()

	relayerDir := os.Getenv("RELAYER_DIR")
	if relayerDir == "" {
		panic(fmt.Errorf("must define RELAYER_DIR env var"))
	}
	relayerDir = utils.ExpandHome(relayerDir)
	if !utils.DirectoryExists(relayerDir) {
		panic(fmt.Errorf("relayer directory %q does not exits", relayerDir))
	}

	chain1RelayerKey, err := key.NewSoft()
	if err != nil {
		panic(err)
	}
	chain2RelayerKey, err := key.NewSoft()
	if err != nil {
		panic(err)
	}

	chain1SubnetID, err := ids.FromString(os.Getenv("CHAIN1_SUBNET_ID"))
	if err != nil {
		panic(err)
	}
	chain1BlockchainID, err := ids.FromString(os.Getenv("CHAIN1_BLOCKCHAIN_ID"))
	if err != nil {
		panic(err)
	}
	chain2SubnetID, err := ids.FromString(os.Getenv("CHAIN2_SUBNET_ID"))
	if err != nil {
		panic(err)
	}
	chain2BlockchainID, err := ids.FromString(os.Getenv("CHAIN2_BLOCKCHAIN_ID"))
	if err != nil {
		panic(err)
	}

	relayerConfig := relayer.CreateBaseRelayerConfig(
		logging.Info.LowerString(),
		filepath.Join(relayerDir, "storage"),
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
		panic(err)
	}

	if err := relayer.FundRelayer(
		relayerConfig,
		chain1BlockchainID,
		chain1PK,
		nil,
		nil,
	); err != nil {
		panic(err)
	}
	if err := relayer.FundRelayer(
		relayerConfig,
		chain2BlockchainID,
		chain2PK,
		nil,
		nil,
	); err != nil {
		panic(err)
	}

	binPath, err := relayer.InstallLatest(relayerDir, "")
	if err != nil {
		panic(err)
	}

	relayerLogPath := filepath.Join(relayerDir, "log.json")

	pid, err := relayer.Execute(binPath, relayerConfigPath, relayerLogPath, "")
	if err != nil {
		if bs, err := os.ReadFile(relayerLogPath); err == nil {
			fmt.Println(string(bs))
		}
		panic(err)
	}

	fmt.Println(pid)
}
