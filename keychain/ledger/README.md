# Ledger Keychain

This Go package provides a keychain implementation for Avalanche that integrates with [Ledger](https://www.ledger.com/) hardware wallets, enabling secure transaction signing with physical device confirmation.

The Ledger keychain allows users to sign Avalanche transactions using their Ledger hardware wallet, keeping private keys secure on the device while maintaining full compatibility with the Avalanche SDK.

## Features

* Hardware-secured transaction signing via Ledger device
* Support for P-Chain, X-Chain, and C-Chain operations
* Compatible with Avalanche's `keychain.Keychain` and `c.EthKeychain` interfaces
* Automatic key derivation using BIP-44 path (m/44'/9000'/0'/0/n)
* Support for both full transaction signing and hash signing
* Retry logic for transient USB communication errors

## Prerequisites

* Ledger Nano S, Nano S Plus, Nano X, Stax, or Flex device
* [Avalanche app](https://github.com/ava-labs/ledger-avalanche) installed on the Ledger device
* Device must be unlocked and Avalanche app must be open

## Usage Example

### Basic Setup and Subnet Creation

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

func main() {
	// Connect to Ledger device
	// Make sure your Ledger is connected, unlocked, and the Avalanche app is open
	fmt.Println("Connecting to Ledger device...")
	device, err := ledger.New()
	if err != nil {
		log.Fatalf("Failed to connect to Ledger: %v", err)
	}
	defer device.Disconnect()

	// Create keychain using the first address (index 0)
	kc, err := ledger.NewKeychain(device, 1)
	if err != nil {
		log.Fatalf("Failed to create keychain: %v", err)
	}

	// Create wallet on Fuji testnet
	ctx := context.Background()
	wallet, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kc,
		kc,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	// Create subnet with threshold of 1
	subnetOwnerAddrs := kc.Addresses().List()
	owner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     subnetOwnerAddrs,
	}

	fmt.Println("Building subnet creation transaction...")
	fmt.Println("*** Please confirm the transaction on your Ledger device ***")

	// Sign and issue create subnet transaction
	signCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	createSubnetTx, err := wallet.P().IssueCreateSubnetTx(
		owner,
		common.WithContext(signCtx),
	)
	if err != nil {
		log.Fatalf("Failed to create subnet: %v", err)
	}

	fmt.Printf("✓ Subnet created with ID: %s\n", createSubnetTx.ID())
}
```

### Using Specific Address Indices

```go
// Create keychain with specific indices (e.g., addresses at indices 0, 5, and 10)
kc, err := ledger.NewKeychainFromIndices(device, []uint32{0, 5, 10})
if err != nil {
	log.Fatalf("Failed to create keychain: %v", err)
}

// Get all addresses managed by this keychain
addresses := kc.Addresses()
fmt.Printf("Managing %d addresses\n", addresses.Len())
```

### Cross-Chain Transfer (P-Chain to C-Chain)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
)

func main() {
	// Connect to Ledger
	device, err := ledger.New()
	if err != nil {
		log.Fatalf("Failed to connect to Ledger: %v", err)
	}
	defer device.Disconnect()

	// Create keychain
	kc, err := ledger.NewKeychain(device, 1)
	if err != nil {
		log.Fatalf("Failed to create keychain: %v", err)
	}

	// Create wallet
	ctx := context.Background()
	wallet, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kc,
		kc,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	// Get chain and asset IDs
	cChainID := wallet.C().Builder().Context().BlockchainID
	avaxAssetID := wallet.P().Builder().Context().AVAXAssetID

	// Export 0.5 AVAX from P-Chain to C-Chain
	exportOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     kc.Addresses().List(),
	}

	fmt.Println("Exporting from P-Chain...")
	fmt.Println("*** Please confirm the transaction on your Ledger device ***")

	exportTx, err := wallet.P().IssueExportTx(
		cChainID,
		[]*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt:          units.Avax / 2, // 0.5 AVAX
				OutputOwners: *exportOwner,
			},
		}},
		common.WithContext(ctx),
	)
	if err != nil {
		log.Fatalf("Failed to export from P-Chain: %v", err)
	}

	fmt.Printf("✓ Exported from P-Chain, tx ID: %s\n", exportTx.ID())

	// Wait for export to be processed
	time.Sleep(10 * time.Second)

	// Import on C-Chain
	// Get Ethereum address from the Ledger-derived public key
	pubKey, err := device.PubKey(0)
	if err != nil {
		log.Fatalf("Failed to get public key: %v", err)
	}
	cChainAddr := pubKey.EthAddress()

	fmt.Println("Importing on C-Chain...")
	fmt.Println("*** Please confirm the transaction on your Ledger device ***")

	importTx, err := wallet.C().IssueImportTx(
		constants.PlatformChainID,
		cChainAddr,
		common.WithContext(ctx),
	)
	if err != nil {
		log.Fatalf("Failed to import on C-Chain: %v", err)
	}

	fmt.Printf("✓ Imported on C-Chain, tx ID: %s\n", importTx.ID())
}
```

## Comprehensive Example

For a complete example demonstrating all supported transaction types across P-Chain, X-Chain, and C-Chain, see [examples/validate-ledger-txs.go](examples/validate-ledger-txs.go).

This example includes:
- All P-Chain transaction types (CreateSubnet, ConvertSubnetToL1, AddValidator, etc.)
- X-Chain transactions (BaseTx, CreateAsset, Import/Export)
- C-Chain atomic transactions (Import/Export)
- L1 validator operations (Register, SetWeight, IncreaseBalance, Disable)

## Integration with Avalanche Wallet

The Ledger keychain implements both `keychain.Keychain` and `c.EthKeychain` interfaces, making it compatible with the Avalanche primary wallet:

```go
wallet, err := primary.MakeWallet(
	ctx,
	endpoint,
	ledgerKeychain,  // For P/X chain operations
	ledgerKeychain,  // For C chain operations (EthKeychain)
	primary.WalletConfig{},
)
```

## Device Connection

The `New()` function automatically:
1. Detects connected Ledger devices
2. Connects to the Avalanche app
3. Returns a `Device` that implements the `Ledger` interface

Always remember to call `Disconnect()` when done:
```go
device, err := ledger.New()
if err != nil {
	log.Fatal(err)
}
defer device.Disconnect()
```
