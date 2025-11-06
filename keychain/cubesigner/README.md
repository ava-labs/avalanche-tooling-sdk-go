# CubeSigner Keychain

This Go package provides a keychain implementation for Avalanche that integrates with [CubeSigner](https://www.cubist.dev/), a remote signing service for secure key management.

CubeSigner enables secure transaction signing without exposing private keys locally, supporting both Avalanche native chains (P-Chain, X-Chain, C-Chain) and Ethereum-compatible chains.

## Features

* Remote transaction signing via CubeSigner API
* Support for P-Chain, X-Chain, and C-Chain operations
* Compatible with Avalanche's `keychain.Keychain` and `c.EthKeychain` interfaces
* Automatic key validation and format conversion
* Support for both hash signing and serialized transaction signing

## Quick Start

For complete usage examples, see:
- [Creating a Subnet](../../examples/cubesigner/create-subnet.go)
- [Cross-Chain Transfer (P-Chain to C-Chain)](../../examples/cubesigner/cross-chain-transfer.go)

## Key Features

### Supported Key Types

The keychain supports CubeSigner keys with the following types:
- `SecpAvaAddr` - Avalanche mainnet secp256k1 keys
- `SecpAvaTestAddr` - Avalanche testnet secp256k1 keys
- `SecpEthAddr` - Ethereum secp256k1 keys

### Signing Methods

#### 1. Serialized Transaction Signing (Recommended)

Used automatically by the Avalanche wallet for P-Chain, X-Chain, and C-Chain transactions. The chain type and network ID are automatically detected from the transaction bytes:

```go
signer, _ := kc.Get(address)
signature, err := signer.Sign(unsignedTxBytes)
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

**Examples:** See [examples/cubesigner](../../examples/cubesigner/) for runnable code examples

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
