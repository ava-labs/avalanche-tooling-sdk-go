// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package local

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
)

// LocalWallet represents a local wallet implementation
type LocalWallet struct {
	wallet         *primary.Wallet            // Primary wallet for P/X/C operations
	accounts       map[string]account.Account // Named accounts map
	activeAccount  string                     // Currently active account name
	defaultNetwork network.Network            // Default network for operations
	seenSubnetIDs  []ids.ID                   // Subnet IDs seen for active account
	evmClient      *ethclient.Client          // EVM client for current chain
	evmRPC         string                     // Current EVM chain RPC URL
}

// Ensure LocalWallet implements Wallet interface
var _ wallet.Wallet = (*LocalWallet)(nil)

// NewLocalWallet creates a new local wallet with the specified network
func NewLocalWallet(net network.Network) (*LocalWallet, error) {
	return &LocalWallet{
		wallet:         nil,
		accounts:       make(map[string]account.Account),
		activeAccount:  "",
		defaultNetwork: net,
		seenSubnetIDs:  []ids.ID{},
	}, nil
}

// SetNetwork sets the default network for wallet operations
func (w *LocalWallet) SetNetwork(net network.Network) {
	w.defaultNetwork = net
}

// Network returns the default network for wallet operations
func (w *LocalWallet) Network() network.Network {
	return w.defaultNetwork
}

// Primary returns the interface for P/X/C chain operations
func (w *LocalWallet) Primary() wallet.PrimaryOperations {
	return &primaryOperations{localWallet: w}
}
