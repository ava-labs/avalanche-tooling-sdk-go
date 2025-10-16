// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
)

// Account represents the interface for different account implementations
type Account interface {
	// GetPChainAddress returns the P-Chain address for the given network
	GetPChainAddress(network network.Network) (string, error)

	GetKeychain() (*secp256k1fx.Keychain, error)
}
