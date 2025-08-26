// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"golang.org/x/exp/maps"
)

type Account struct {
	keychain.Keychain
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
