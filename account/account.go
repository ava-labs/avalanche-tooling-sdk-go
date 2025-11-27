// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
)

// Account represents a set of X/P/C-Chain and EVM addresses derived from a single private key.
// This is equivalent to an avalanchego secp256k1fx.Keychain with a single key that can generate
// addresses for all chain types (AVAX-type addresses for X/P chains and EVM-type for C-Chain).
//
// A LocalAccount (which implements Account) can be created using NewLocalAccount(), or imported
// from an existing private key using ImportFromPrivateKey() or ImportFromFile(). Accounts can be
// imported into a wallet for transaction operations.
type Account interface {
	// Name returns the account name
	Name() string

	// GetPChainAddress returns the P-Chain address for the given network
	GetPChainAddress(network network.Network) (string, error)

	// GetXChainAddress returns the X-Chain address for the given network
	GetXChainAddress(network network.Network) (string, error)

	// GetCChainAddress returns the C-Chain address (same as EVM address)
	GetCChainAddress() (string, error)

	// GetEVMAddress returns the EVM address (0x format)
	GetEVMAddress() (string, error)

	// GetKeychain returns the underlying secp256k1fx.Keychain for signing operations
	GetKeychain() (*secp256k1fx.Keychain, error)
}
