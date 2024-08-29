// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/icm"
)

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	icmVersion, err := icm.GetLatestVersion()
	if err != nil {
		fatal(err)
	}
	td := icm.Deployer{}
	if err := td.DownloadAssets(icmVersion); err != nil {
		fatal(err)
	}
	chain1MessengerAlreadyDeployed, chain1MessengerAddress, chain1RegistryAddress, err := td.Deploy(
		os.Getenv("CHAIN1_RPC"),
		os.Getenv("CHAIN1_PK"),
		true,
		true,
	)
	if err != nil {
		fatal(err)
	}
	chain2MessengerAlreadyDeployed, chain2MessengerAddress, chain2RegistryAddress, err := td.Deploy(
		os.Getenv("CHAIN2_RPC"),
		os.Getenv("CHAIN2_PK"),
		true,
		true,
	)
	if err != nil {
		fatal(err)
	}
	chain1RegistryAddress = "0x4bC756894C6CB10A5735816E25132486F5a1cE8f"
	chain2RegistryAddress = "0x302a91b43d974Cd6f12f4Eae8cADBc8efB7359c8"
	fmt.Println(chain1MessengerAlreadyDeployed)
	fmt.Println(chain1MessengerAddress)
	fmt.Println(chain1RegistryAddress)
	fmt.Println(chain2MessengerAlreadyDeployed)
	fmt.Println(chain2MessengerAddress)
	fmt.Println(chain2RegistryAddress)

}
