// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package keychain

import (
	"avalanche-tooling-sdk-go/avalanche"

	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
)

type Keychain struct {
	keychain.Keychain
	Network       avalanche.Network
	Ledger        keychain.Ledger
	UsesLedger    bool
	LedgerIndices []uint32
}
