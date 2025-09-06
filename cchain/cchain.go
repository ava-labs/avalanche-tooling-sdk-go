// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cchain

import (
	"github.com/ava-labs/coreth/plugin/evm/atomic"
)

// TxFromBytes attempts to unmarshal bytes into a C-Chain atomic transaction.
// Returns the transaction and true if successful, nil and false otherwise.
func TxFromBytes(b []byte) (*atomic.Tx, bool) {
	if len(b) == 0 {
		return nil, false
	}

	tx := new(atomic.Tx)
	if _, err := atomic.Codec.Unmarshal(b, tx); err != nil {
		return nil, false
	}

	return tx, true
}

// IsCChainTx checks if the provided bytes represent a C-Chain atomic transaction.
func IsCChainTx(b []byte) bool {
	_, ok := TxFromBytes(b)
	return ok
}
