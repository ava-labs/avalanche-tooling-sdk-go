// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
)

// LocalAccount represents a local account implementation
type LocalAccount struct {
	softKey *key.SoftKey
}

// NewLocalAccount creates a new LocalAccount
func NewLocalAccount() (Account, error) {
	k, err := key.NewSoft()
	if err != nil {
		return nil, err
	}
	return &LocalAccount{
		softKey: k,
	}, nil
}

// ImportFromString imports an account from a hex-encoded private key string
func ImportFromString(privateKeyHex string) (Account, error) {
	k, err := key.LoadSoftFromBytes([]byte(privateKeyHex))
	if err != nil {
		return nil, err
	}
	return &LocalAccount{
		softKey: k,
	}, nil
}

// ImportFromPath imports an account from a file path
func ImportFromPath(keyPath string) (Account, error) {
	kb, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return ImportFromString(string(kb))
}

func (a *LocalAccount) GetPChainAddress(network network.Network) (string, error) {
	if a.softKey == nil {
		return "", fmt.Errorf("softKey not initialized")
	}
	pchainAddrs, err := a.softKey.GetNetworkChainAddress(network, "P")
	return pchainAddrs[0], err
}

func (a *LocalAccount) GetXChainAddress(network network.Network) (string, error) {
	if a.softKey == nil {
		return "", fmt.Errorf("softKey not initialized")
	}
	xchainAddrs, err := a.softKey.GetNetworkChainAddress(network, "X")
	return xchainAddrs[0], err
}

func (a *LocalAccount) GetCChainAddress(network network.Network) (string, error) {
	if a.softKey == nil {
		return "", fmt.Errorf("softKey not initialized")
	}
	cchainAddrs, err := a.softKey.GetNetworkChainAddress(network, "C")
	return cchainAddrs[0], err
}

func (a *LocalAccount) GetEVMAddress() (string, error) {
	if a.softKey == nil {
		return "", fmt.Errorf("softKey not initialized")
	}
	return a.softKey.C(), nil
}

func (a *LocalAccount) GetKeychain() (*secp256k1fx.Keychain, error) {
	if a.softKey == nil {
		return nil, fmt.Errorf("softKey not initialized")
	}
	return a.softKey.KeyChain(), nil
}

// PrivateKey exports the private key in hex format
func (a *LocalAccount) PrivateKey() (string, error) {
	if a.softKey == nil {
		return "", fmt.Errorf("softKey not initialized")
	}
	return a.softKey.PrivKeyHex(), nil
}
