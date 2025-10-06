// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
)

// Account represents the interface for different account implementations
// in the Avalanche ecosystem. An account provides a unified interface
// for managing cryptographic identities across various Avalanche networks
// and chains. It abstracts the underlying key management implementation,
// allowing for different account types (software wallets, hardware wallets,
// server wallets etc.).
type Account struct {
	*key.SoftKey
}

func NewAccount() (Account, error) {
	k, err := key.NewSoft()
	if err != nil {
		return Account{}, err
	}
	return Account{
		SoftKey: k,
	}, nil
}

func Import(keyPath string) (Account, error) {
	k, err := key.LoadSoft(keyPath)
	if err != nil {
		return Account{}, err
	}
	return Account{
		SoftKey: k,
	}, nil
}

func (a *Account) GetPChainAddress(network network.Network) (string, error) {
	if a.SoftKey == nil {
		return "", fmt.Errorf("SoftKey not initialized")
	}
	pchainAddrs, err := a.SoftKey.GetNetworkChainAddress(network, "P")
	return pchainAddrs[0], err
}
