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

## Usage Examples

Working examples are available in the [examples/ledger](../../examples/ledger/) directory:

### Basic Examples

* **[create-subnet](../../examples/ledger/create-subnet.go)** - Create a subnet using a Ledger device
  - Shows basic Ledger connection and keychain setup
  - Demonstrates P-Chain subnet creation with Ledger signing

* **[cross-chain-transfer](../../examples/ledger/cross-chain-transfer.go)** - Transfer AVAX between chains
  - Export AVAX from P-Chain to C-Chain
  - Import AVAX on C-Chain
  - Derive Ethereum address from Ledger public key

### Comprehensive Example

* **[validate-ledger-txs](../../examples/ledger/validate-ledger-txs.go)** - Complete demonstration of all transaction types
  - All P-Chain transaction types (CreateSubnet, ConvertSubnetToL1, AddValidator, etc.)
  - X-Chain transactions (BaseTx, CreateAsset, Import/Export)
  - C-Chain atomic transactions (Import/Export)
  - L1 validator operations (Register, SetWeight, IncreaseBalance, Disable)

### Quick Start

```go
// Connect to Ledger device
device, err := ledger.New()
if err != nil {
    log.Fatal(err)
}
defer device.Disconnect()

// Create keychain with first address
kc, err := ledger.NewKeychain(device, 1)

// Or use specific address indices
kc, err := ledger.NewKeychainFromIndices(device, []uint32{0, 5, 10})
```

## Package Structure

```
keychain/ledger/
├── ledger.go                    # Ledger interface definition
├── ledger_device.go             # Device implementation
├── ledger_device_test.go        # Tests for device implementation
├── ledger_keychain.go           # KeyChain and signer implementations
├── ledger_keychain_test.go      # Tests for keychain implementation
├── mocks_generate_test.go       # Mock generation directives
├── ledgermock/
│   └── ledger.go                # Generated mock for Ledger interface
└── README.md                    # This file
```

**Core Files:**
- **ledger_device.go**: Implements the `Ledger` interface using `github.com/ava-labs/ledger-avalanche-go`
- **ledger_keychain.go**: Implements `keychain.Keychain` and `c.EthKeychain` interfaces, manages address derivation and transaction signing

**Examples:** See [examples/ledger](../../examples/ledger/) for runnable code examples

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
