// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"

	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
)

// LocalAccount represents a local account implementation
type LocalAccount struct {
	*key.SoftKey
}

// NewLocalAccount creates a new LocalAccount
func NewLocalAccount() (Account, error) {
	k, err := key.NewSoft()
	if err != nil {
		return nil, err
	}
	return &LocalAccount{
		SoftKey: k,
	}, nil
}

func Import(keyPath string) (Account, error) {
	k, err := key.LoadSoft(keyPath)
	if err != nil {
		return nil, err
	}
	return &LocalAccount{
		SoftKey: k,
	}, nil
}

func (a *LocalAccount) GetPChainAddress(network network.Network) (string, error) {
	if a.SoftKey == nil {
		return "", fmt.Errorf("SoftKey not initialized")
	}
	pchainAddrs, err := a.SoftKey.GetNetworkChainAddress(network, "P")
	return pchainAddrs[0], err
}

func (a *LocalAccount) GetKeychain() (*secp256k1fx.Keychain, error) {
	if a.SoftKey == nil {
		return nil, fmt.Errorf("SoftKey not initialized")
	}
	return a.SoftKey.KeyChain(), nil
}
