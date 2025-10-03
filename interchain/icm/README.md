# Interchain Messaging (ICM) SDK

This Go package provides utilities for deploying and interacting with Interchain Messaging (ICM) contracts on Avalanche L1s.

ICM (formerly known as Teleporter) enables asynchronous cross-chain communication between EVM-based Avalanche L1s using the Avalanche Warp Messaging protocol.

## Features

* Deploy TeleporterMessenger contracts using Nick's method
* Deploy TeleporterRegistry contracts
* Add ICM contracts directly to genesis allocations
* Load ICM contract deployment assets from default embedded versions, GitHub releases, or local files
* Automatic deployer funding when needed
* Support for custom ICM contract versions

## Usage Example

```go
// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/icm"
)

func main() {
	// Create a new ICM deployer
	deployer := &icm.Deployer{}

	// Load default ICM v1.0.0 deployment assets (embedded)
	deployer.LoadDefault()

	// Alternatively, load from a specific GitHub release:
	// if err := deployer.LoadFromRelease("v1.0.0", ""); err != nil {
	//     fmt.Printf("Failed to load from release: %v\n", err)
	//     return
	// }

	// Or load from local files:
	// if err := deployer.LoadFromFiles(
	//     "/path/to/messenger_contract_address.txt",
	//     "/path/to/messenger_deployer_address.txt",
	//     "/path/to/messenger_deployment_transaction.txt",
	//     "/path/to/registry_bytecode.txt",
	// ); err != nil {
	//     fmt.Printf("Failed to load from files: %v\n", err)
	//     return
	// }

	// Deploy both TeleporterMessenger and TeleporterRegistry
	rpcURL := "http://localhost:9650/ext/bc/myblockchainid/rpc"
	privateKey := "your-private-key-hex"

	messengerAddr, registryAddr, err := deployer.Deploy(rpcURL, privateKey)
	if err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
		return
	}

	fmt.Printf("TeleporterMessenger deployed at: %s\n", messengerAddr)
	fmt.Printf("TeleporterRegistry deployed at: %s\n", registryAddr)

	// You can also deploy contracts individually:
	// messengerAddr, err := deployer.DeployMessenger(rpcURL, privateKey)
	// registryAddr, err := deployer.DeployRegistry(rpcURL, privateKey)
}
```

## Adding ICM Contracts to Genesis

You can pre-deploy ICM contracts in your L1's genesis configuration:

```go
package main

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/icm"
	"github.com/ava-labs/subnet-evm/core"
)

func main() {
	// Create genesis allocations
	allocs := core.GenesisAlloc{}

	// Add TeleporterMessenger contract to genesis
	// This deploys the contract at: 0x253b2784c75e510dD0fF1da844684a1aC0aa5fcf
	icm.AddMessengerContractToAllocations(allocs)

	// Add TeleporterRegistry contract to genesis
	// This deploys the contract at: 0xF86Cb19Ad8405AEFa7d09C778215D2Cb6eBfB228
	if err := icm.AddRegistryContractToAllocations(allocs); err != nil {
		fmt.Printf("Failed to add registry to genesis: %v\n", err)
		return
	}

	// Use allocs in your genesis configuration
	fmt.Printf("ICM contracts added to genesis\n")
}
```

## Deployment Details

### TeleporterMessenger

The TeleporterMessenger contract is deployed using Nick's method, which ensures a deterministic contract address (`0x253b2784c75e510dD0fF1da844684a1aC0aa5fcf`) across all chains. The deployer automatically funds the deployer address with the required 10 AVAX if needed.

### TeleporterRegistry

The TeleporterRegistry contract is deployed as a standard contract and initialized with the TeleporterMessenger contract address at version 1.

When added to genesis, the registry is deployed at `0xF86Cb19Ad8405AEFa7d09C778215D2Cb6eBfB228`.

## Error Handling

If the TeleporterMessenger is already deployed on the chain, `DeployMessenger` and `Deploy` will return `icm.ErrMessengerAlreadyDeployed`. The messenger address is still returned in this case.
