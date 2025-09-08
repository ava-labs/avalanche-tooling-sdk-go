// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
)

type Account struct {
	*key.SoftKey
	*keychain.Keychain
}

func NewAccount() (Account, error) {
	k, err := key.NewSoft(0)
	if err != nil {
		return Account{}, err
	}
	return Account{
		SoftKey:  k,
		Keychain: nil, // Will be set later if needed
	}, nil
}

func (a *Account) GetPChainAddress() (string, error) {
	if a.SoftKey == nil {
		return "", fmt.Errorf("SoftKey not initialized")
	}
	cChainAddr := a.SoftKey.C()
	return cChainAddr, nil
}
