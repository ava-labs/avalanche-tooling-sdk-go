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

	// GetXChainAddress returns the X-Chain address for the given network
	GetXChainAddress(network network.Network) (string, error)

	// GetCChainAddress returns the C-Chain address for the given network
	GetCChainAddress(network network.Network) (string, error)

	// GetEVMAddress returns the EVM address (0x format)
	GetEVMAddress() (string, error)

	GetKeychain() (*secp256k1fx.Keychain, error)
}

// AccountInfo contains public information about an account
// Does NOT expose private key material
type AccountInfo struct {
	Name       string // Account name
	PAddress   string // P-Chain address
	XAddress   string // X-Chain address
	CAddress   string // C-Chain address
	EVMAddress string // EVM address (0x format)
}

// AccountSpec contains account creation/import specifications
type AccountSpec struct {
	PrivateKey     string // Hex-encoded private key (for local accounts)
	DerivationPath string // For Ledger accounts
	RemoteKeyID    string // For remote signing services
}
