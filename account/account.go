// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
)

type Account struct {
	*key.SoftKey
	*keychain.Keychain
}

func NewAccount() (Account, error) {
	k, err := key.NewSoft()
	if err != nil {
		return Account{}, err
	}
	return Account{
		SoftKey:  k,
		Keychain: nil, // Will be set later if needed
	}, nil
}

func (a *Account) GetPChainAddress(network network.Network) (string, error) {
	if a.SoftKey == nil {
		return "", fmt.Errorf("SoftKey not initialized")
	}
	pchainAddrs, err := a.SoftKey.GetNetworkChainAddress(network, "P")
	return pchainAddrs[0], err
}
