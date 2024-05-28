// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package keychain

import (
	"fmt"

	"avalanche-tooling-sdk-go/avalanche"
	"avalanche-tooling-sdk-go/key"
	"avalanche-tooling-sdk-go/ledger"
	"avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanchego/utils/crypto/keychain"

	"golang.org/x/exp/maps"
)

type Keychain struct {
	keychain.Keychain
	network       avalanche.Network
	ledgerDevice  *ledger.LedgerDevice
	ledgerIndices []uint32
}

func (kc *Keychain) P() ([]string, error) {
	return utils.P(kc.network.HRP(), kc.Addresses().List())
}

func (kc *Keychain) LedgerEnabled() bool {
	return kc.ledgerDevice != nil
}

func (kc *Keychain) AddLedgerIndices(indices []uint32) error {
	if kc.LedgerEnabled() {
		kc.ledgerIndices = utils.Unique(append(kc.ledgerIndices, indices...))
		utils.Uint32Sort(kc.ledgerIndices)
		newKc, err := keychain.NewLedgerKeychainFromIndices(kc.ledgerDevice, kc.ledgerIndices)
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
		indices, err := kc.ledgerDevice.FindAddresses(addresses, 0)
		if err != nil {
			return err
		}
		return kc.AddLedgerIndices(maps.Values(indices))
	}
	return fmt.Errorf("keychain is not ledger enabled")
}

func (kc *Keychain) AddLedgerFunds(amount uint64) error {
	if kc.LedgerEnabled() {
		indices, err := kc.ledgerDevice.FindFunds(kc.network, amount, 0)
		if err != nil {
			return err
		}
		return kc.AddLedgerIndices(indices)
	}
	return fmt.Errorf("keychain is not ledger enabled")
}

func NewKeychain(
	network avalanche.Network,
	keyPath string,
	useEwoq bool,
	useLedger bool,
	ledgerAddresses []string,
	requiredFunds uint64,
) (*Keychain, error) {
	// get keychain accessor
	if useLedger {
		dev, err := ledger.New()
		if err != nil {
			return nil, err
		}
		kc := Keychain{
			ledgerDevice: dev,
			network:      network,
		}
		if err := kc.AddLedgerIndices([]uint32{0}); err != nil {
			return nil, err
		}
		if requiredFunds > 0 {
			if err := kc.AddLedgerFunds(requiredFunds); err != nil {
				return nil, err
			}
		}
		if len(ledgerAddresses) > 0 {
			if err := kc.AddLedgerAddresses(ledgerAddresses); err != nil {
				return nil, err
			}
		}
		return &kc, nil
	}
	if useEwoq {
		sf, err := key.LoadEwoq()
		if err != nil {
			return nil, err
		}
		kc := Keychain{
			Keychain: sf.KeyChain(),
			network:  network,
		}
		return &kc, nil
	}
	if keyPath != "" {
		sf, err := key.LoadSoft(keyPath)
		if err != nil {
			return nil, err
		}
		kc := Keychain{
			Keychain: sf.KeyChain(),
			network:  network,
		}
		return &kc, nil
	}
	return nil, fmt.Errorf("not keychain option defined")
}
