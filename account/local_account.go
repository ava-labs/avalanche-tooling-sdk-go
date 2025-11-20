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
	name    string
	softKey *key.SoftKey
}

// NewLocalAccount creates a new local account with a freshly generated private key
func NewLocalAccount(name string) (*LocalAccount, error) {
	k, err := key.NewSoft()
	if err != nil {
		return nil, err
	}
	return &LocalAccount{
		name:    name,
		softKey: k,
	}, nil
}

// ImportFromPrivateKey imports a local account from a hex-encoded private key string
func ImportFromPrivateKey(name string, privateKeyHex string) (*LocalAccount, error) {
	k, err := key.LoadSoftFromBytes([]byte(privateKeyHex))
	if err != nil {
		return nil, err
	}
	return &LocalAccount{
		name:    name,
		softKey: k,
	}, nil
}

// ImportFromFile imports a local account from a file path containing a private key
func ImportFromFile(name string, keyPath string) (*LocalAccount, error) {
	kb, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return ImportFromPrivateKey(name, string(kb))
}

func (a *LocalAccount) Name() string {
	return a.name
}

func (a *LocalAccount) GetPChainAddress(network network.Network) (string, error) {
	if a == nil || a.softKey == nil {
		return "", fmt.Errorf("uninitialized local account")
	}
	pchainAddrs, err := a.softKey.GetNetworkChainAddress(network, "P")
	if err != nil {
		return "", err
	}
	if len(pchainAddrs) == 0 {
		return "", fmt.Errorf("no P-Chain address found")
	}
	return pchainAddrs[0], nil
}

func (a *LocalAccount) GetXChainAddress(network network.Network) (string, error) {
	if a == nil || a.softKey == nil {
		return "", fmt.Errorf("uninitialized local account")
	}
	xchainAddrs, err := a.softKey.GetNetworkChainAddress(network, "X")
	if err != nil {
		return "", err
	}
	if len(xchainAddrs) == 0 {
		return "", fmt.Errorf("no X-Chain address found")
	}
	return xchainAddrs[0], nil
}

func (a *LocalAccount) GetCChainAddress() (string, error) {
	if a == nil || a.softKey == nil {
		return "", fmt.Errorf("uninitialized local account")
	}
	// C-Chain uses EVM address format (0x...)
	return a.softKey.C(), nil
}

func (a *LocalAccount) GetEVMAddress() (string, error) {
	if a == nil || a.softKey == nil {
		return "", fmt.Errorf("uninitialized local account")
	}
	return a.softKey.C(), nil
}

func (a *LocalAccount) GetKeychain() (*secp256k1fx.Keychain, error) {
	if a == nil || a.softKey == nil {
		return nil, fmt.Errorf("uninitialized local account")
	}
	return a.softKey.KeyChain(), nil
}

// PrivateKey exports the private key in hex format
func (a *LocalAccount) PrivateKey() (string, error) {
	if a == nil || a.softKey == nil {
		return "", fmt.Errorf("uninitialized local account")
	}
	return a.softKey.PrivKeyHex(), nil
}
