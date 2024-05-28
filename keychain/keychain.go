// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package keychain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/utils/crypto/keychain"

	"avalanche-tooling-sdk-go/avalanche"
	"avalanche-tooling-sdk-go/ledger"
	"avalanche-tooling-sdk-go/utils"

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

func GetKeychain(
	app *application.Avalanche,
	useEwoq bool,
	useLedger bool,
	ledgerAddresses []string,
	network models.Network,
	requiredFunds uint64,
) (*Keychain, error) {
	// get keychain accessor
	if useLedger {
		ledgerDevice, err := ledger.New()
		if err != nil {
			return nil, err
		}
		// always have index 0, for change
		ledgerIndices := []uint32{0}
		if requiredFunds > 0 {
			ledgerIndicesAux, err := searchForFundedLedgerIndices(network, ledgerDevice, requiredFunds)
			if err != nil {
				return nil, err
			}
			ledgerIndices = append(ledgerIndices, ledgerIndicesAux...)
		}
		if len(ledgerAddresses) > 0 {
			ledgerIndicesAux, err := getLedgerIndices(ledgerDevice, ledgerAddresses)
			if err != nil {
				return nil, err
			}
			ledgerIndices = append(ledgerIndices, ledgerIndicesAux...)
		}
		ledgerIndicesSet := set.Set[uint32]{}
		ledgerIndicesSet.Add(ledgerIndices...)
		ledgerIndices = ledgerIndicesSet.List()
		utils.SortUint32(ledgerIndices)
		if err := showLedgerAddresses(network, ledgerDevice, ledgerIndices); err != nil {
			return nil, err
		}
		kc, err := keychain.NewLedgerKeychainFromIndices(ledgerDevice, ledgerIndices)
		if err != nil {
			return nil, err
		}
		return NewKeychain(network, kc, ledgerDevice, ledgerIndices), nil
	}
	if useEwoq {
		sf, err := app.GetKey("ewoq", network, false)
		if err != nil {
			return nil, err
		}
		kc := sf.KeyChain()
		return NewKeychain(network, kc, nil, nil), nil
	}
	sf, err := app.GetKey(keyName, network, false)
	if err != nil {
		return nil, err
	}
	kc := sf.KeyChain()
	return NewKeychain(network, kc, nil, nil), nil
}
