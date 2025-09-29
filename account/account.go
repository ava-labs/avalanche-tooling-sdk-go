// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanchego/ids"
)

// Account represents the interface for different account implementations
type Account interface {
	// Addresses returns all addresses associated with this account
	Addresses() []ids.ShortID
	// GetPChainAddress returns the P-Chain address for the given network
	GetPChainAddress(network network.Network) (string, error)
	// NewAccount creates a new account of the same type
	NewAccount() (Account, error)

	GetKeychain() (*secp256k1fx.Keychain, error)
}
