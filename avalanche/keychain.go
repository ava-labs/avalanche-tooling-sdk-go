// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import "github.com/ava-labs/avalanchego/utils/crypto/keychain"

type Keychain struct {
	Network Network

	Keychain keychain.Keychain

	Ledger keychain.Ledger

	UsesLedger bool

	LedgerIndices []uint32
}
