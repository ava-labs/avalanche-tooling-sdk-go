// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package keychain

import (
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/ledger"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"golang.org/x/exp/maps"
)

type Keychain struct {
	keychain.Keychain
	network avalanche.Network
	Ledger  *Ledger
}

// TODO: add descriptions for the properties below
type LedgerParams struct {
	RequiredFunds   uint64
	LedgerAddresses []string
}

// TODO: add descriptions for the properties below
type Ledger struct {
	LedgerDevice  *ledger.LedgerDevice
	LedgerIndices []uint32
}

// NewKeychain will generate a new key pair in the provided keyPath if no .pk file currently
// exists in the provided keyPath
func NewKeychain(
	network avalanche.Network,
	keyPath string,
	ledgerInfo *LedgerParams,
) (*Keychain, error) {
	if ledgerInfo != nil {
		if keyPath != "" {
			return nil, fmt.Errorf("keychain can only created either from key path or ledger, not both")
		}
		dev, err := ledger.New()
		if err != nil {
			return nil, err
		}
		kc := Keychain{
			Ledger: &Ledger{
				LedgerDevice: dev,
			},
			network: network,
		}
		if err := kc.AddLedgerIndices([]uint32{0}); err != nil {
			return nil, err
		}
		if ledgerInfo.RequiredFunds > 0 {
			if err := kc.AddLedgerFunds(ledgerInfo.RequiredFunds); err != nil {
				return nil, err
			}
		}
		if len(ledgerInfo.LedgerAddresses) > 0 {
			if err := kc.AddLedgerAddresses(ledgerInfo.LedgerAddresses); err != nil {
				return nil, err
			}
		}
		return &kc, nil
	}
	sf, err := key.LoadSoftOrCreate(keyPath)
	if err != nil {
		return nil, err
	}
	kc := Keychain{
		Keychain: sf.KeyChain(),
		network:  network,
	}
	return &kc, nil
}

// P returns string formatted addresses in the keychain
func (kc *Keychain) P() ([]string, error) {
	return utils.P(kc.network.HRP(), kc.Addresses().List())
}

func (kc *Keychain) LedgerEnabled() bool {
	return kc.Ledger.LedgerDevice != nil
}

func (kc *Keychain) AddLedgerIndices(indices []uint32) error {
	if kc.LedgerEnabled() {
		kc.Ledger.LedgerIndices = utils.Unique(append(kc.Ledger.LedgerIndices, indices...))
		utils.Uint32Sort(kc.Ledger.LedgerIndices)
		newKc, err := keychain.NewLedgerKeychainFromIndices(kc.Ledger.LedgerDevice, kc.Ledger.LedgerIndices)
		if err != nil {
			return err
		}
		kc.Keychain = newKc
		return nil
	}
	return fmt.Errorf("keychain is not ledger enabled")
}

func (kc *Keychain) AddLedgerAddresses(addresses []string) error {
	if kc.LedgerEnabled() {
		indices, err := kc.Ledger.LedgerDevice.FindAddresses(addresses, 0)
		if err != nil {
			return err
		}
		return kc.AddLedgerIndices(maps.Values(indices))
	}
	return fmt.Errorf("keychain is not ledger enabled")
}

func (kc *Keychain) AddLedgerFunds(amount uint64) error {
	if kc.LedgerEnabled() {
		indices, err := kc.Ledger.LedgerDevice.FindFunds(kc.network, amount, 0)
		if err != nil {
			return err
		}
		return kc.AddLedgerIndices(indices)
	}
	return fmt.Errorf("keychain is not ledger enabled")
}
