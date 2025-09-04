// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
)

type Account struct {
	*keychain.SoftKey
	*keychain.Keychain
}

func NewAccount() (Account, error) {
	k, err := key.NewSoft(0)
	if err != nil {
		return err
	}
	return nil, nil
}

func (a *Account) GetPChainAddress() (string, error) {
	cChainAddr := sk.C()

}
