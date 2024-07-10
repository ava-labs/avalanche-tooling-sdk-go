// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package keychain

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
)

type Keychain struct {
	keychain.Keychain
}

// NewKeychain will generate a new key pair in the provided keyPath if no .pk file currently
// exists in the provided keyPath
func NewKeychain(
	keyPath string,
) (*Keychain, error) {
	sf, err := key.LoadSoftOrCreate(utils.ExpandHome(keyPath))
	if err != nil {
		return nil, err
	}
	kc := Keychain{
		Keychain: sf.KeyChain(),
	}
	return &kc, nil
}

func KeychainFromKey(
	sf *key.SoftKey,
) *Keychain {
	kc := Keychain{
		Keychain: sf.KeyChain(),
	}
	return &kc
}

// P returns string formatted addresses in the keychain
func (kc *Keychain) P(network avalanche.Network) ([]string, error) {
	return utils.P(network.HRP(), kc.Addresses().List())
}
