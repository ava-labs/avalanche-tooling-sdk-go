# CubeSigner Keychain

This Go package provides a keychain implementation for Avalanche that integrates with [CubeSigner](https://www.cubist.dev/), a remote signing service for secure key management.

CubeSigner enables secure transaction signing without exposing private keys locally, supporting both Avalanche native chains (P-Chain, X-Chain, C-Chain) and Ethereum-compatible chains.

## Features

* Remote transaction signing via CubeSigner API
* Support for P-Chain, X-Chain, and C-Chain operations
* Compatible with Avalanche's `keychain.Keychain` and `c.EthKeychain` interfaces
* Automatic key validation and format conversion
* Support for both hash signing and serialized transaction signing

## Usage Example

### Creating a Subnet

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/cubesigner"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/session"
)

func main() {
	// Get key ID from environment
	keyID := os.Getenv("CUBESIGNER_KEY_ID")
	if keyID == "" {
		log.Fatal("CUBESIGNER_KEY_ID environment variable is required")
	}

	// Initialize CubeSigner client
	sessionFile := "session.json"
	manager, err := session.NewJsonSessionManager(&sessionFile)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}

	apiClient, err := client.NewApiClient(manager)
	if err != nil {
		log.Fatalf("Failed to create API client: %v", err)
	}

	// Create CubeSigner keychain
	kc, err := cubesigner.NewKeychain(apiClient, []string{keyID})
	if err != nil {
		log.Fatalf("Failed to create keychain: %v", err)
	}

	// Create primary wallet on Fuji testnet
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

	// Issue create subnet transaction
	createSubnetTx, err := wallet.P().IssueCreateSubnetTx(
		owner,
		common.WithContext(ctx),
	)
	if err != nil {
		log.Fatalf("Failed to create subnet: %v", err)
	}

	fmt.Printf("Subnet created with ID: %s\n", createSubnetTx.ID())
}
```

### Cross-Chain Transfer (P-Chain to C-Chain)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/cubesigner"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/session"
)

func main() {
	// Get key IDs from environment
	pChainKeyID := os.Getenv("PCHAIN_KEY_ID")
	cChainKeyID := os.Getenv("CCHAIN_KEY_ID")
	if pChainKeyID == "" || cChainKeyID == "" {
		log.Fatal("PCHAIN_KEY_ID and CCHAIN_KEY_ID environment variables are required")
	}

	// Initialize CubeSigner client
	sessionFile := "session.json"
	manager, err := session.NewJsonSessionManager(&sessionFile)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}

	apiClient, err := client.NewApiClient(manager)
	if err != nil {
		log.Fatalf("Failed to create API client: %v", err)
	}

	// Create keychains
	kcPChain, err := cubesigner.NewKeychain(apiClient, []string{pChainKeyID})
	if err != nil {
		log.Fatalf("Failed to create P-chain keychain: %v", err)
	}

	kcCChain, err := cubesigner.NewKeychain(apiClient, []string{cChainKeyID})
	if err != nil {
		log.Fatalf("Failed to create C-chain keychain: %v", err)
	}

	// Create wallet for P-chain operations
	ctx := context.Background()
	walletPChain, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kcPChain,
		kcPChain,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to create P-chain wallet: %v", err)
	}

	// Get chain and asset IDs
	cChainID := walletPChain.C().Builder().Context().BlockchainID
	avaxAssetID := walletPChain.P().Builder().Context().AVAXAssetID

	// Export 0.5 AVAX from P-chain to C-chain
	exportOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     kcCChain.Addresses().List(),
	}

	exportTx, err := walletPChain.P().IssueExportTx(
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
		log.Fatalf("Failed to export from P-chain: %v", err)
	}

	fmt.Printf("Exported from P-chain, tx ID: %s\n", exportTx.ID())

	// Wait for export to be processed
	time.Sleep(10 * time.Second)

	// Create wallet for C-chain operations
	walletCChain, err := primary.MakeWallet(
		ctx,
		primary.FujiAPIURI,
		kcCChain,
		kcCChain,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to create C-chain wallet: %v", err)
	}

	// Import on C-chain
	cChainAddr := kcCChain.EthAddresses().List()[0]
	importTx, err := walletCChain.C().IssueImportTx(
		constants.PlatformChainID,
		cChainAddr,
		common.WithContext(ctx),
	)
	if err != nil {
		log.Fatalf("Failed to import on C-chain: %v", err)
	}

	fmt.Printf("Imported on C-chain, tx ID: %s\n", importTx.ID())
}
```

## Key Features

### Supported Key Types

The keychain supports CubeSigner keys with the following types:
- `SecpAvaAddr` - Avalanche mainnet secp256k1 keys
- `SecpAvaTestAddr` - Avalanche testnet secp256k1 keys
- `SecpEthAddr` - Ethereum secp256k1 keys

### Signing Methods

#### 1. Serialized Transaction Signing (Recommended)

Used automatically by the Avalanche wallet for P-Chain and X-Chain transactions. Requires:
- `ChainAlias` option: "P", "X", or "C"
- `NetworkID` option: Network ID

```go
import "github.com/ava-labs/avalanchego/utils/crypto/keychain"

// Example with signing options
signer, _ := kc.Get(address)
signature, err := signer.Sign(
	unsignedTxBytes,
	keychain.WithChainAlias("P"),
	keychain.WithNetworkID(constants.FujiID),
)
```

#### 2. Hash Signing

Used for signing arbitrary data hashed using Keccak-256:

```go
signer, _ := kc.Get(address)
signature, err := signer.SignHash(hashBytes)
```

## Package Structure

```
keychain/cubesigner/
├── cubesigner.go                  # CubeSigner client interface definition
├── cubesigner_keychain.go         # Keychain and signer implementations
├── cubesigner_keychain_test.go    # Tests for keychain implementation
├── mocks_generate_test.go         # Mock generation directives
├── cubesignermock/
│   └── cubesigner_client.go       # Generated mock for CubeSignerClient interface
└── README.md                      # This file
```

**Core Files:**
- **cubesigner_keychain.go**: Implements `keychain.Keychain` and `c.EthKeychain` interfaces, handles key validation, format conversion, and transaction signing via CubeSigner API

## Integration with Avalanche Wallet

The CubeSigner keychain implements both `keychain.Keychain` and `c.EthKeychain` interfaces, making it compatible with the Avalanche primary wallet:

```go
wallet, err := primary.MakeWallet(
	ctx,
	endpoint,
	cubesignerKeychain,  // For P/X chain operations
	cubesignerKeychain,  // For C chain operations
	primary.WalletConfig{},
)
```
