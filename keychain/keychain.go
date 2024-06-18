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
	network avalanche.Network
}

// NewKeychain will generate a new key pair in the provided keyPath if no .pk file currently
// exists in the provided keyPath
func NewKeychain(
	network avalanche.Network,
	keyPath string,
) (*Keychain, error) {
	sf, err := key.LoadSoftOrCreate(utils.ExpandHome(keyPath))
	if err != nil {
		return nil, err
	}
	kc := Keychain{
		Keychain: sf.KeyChain(),
		network:  network,
	}
	return &kc, nil
}

func KeychainFromKey(
	network avalanche.Network,
	sf *key.SoftKey,
) *Keychain {
	kc := Keychain{
		Keychain: sf.KeyChain(),
		network:  network,
	}
	return &kc
}

// P returns string formatted addresses in the keychain
func (kc *Keychain) P() ([]string, error) {
	return utils.P(kc.network.HRP(), kc.Addresses().List())
}
