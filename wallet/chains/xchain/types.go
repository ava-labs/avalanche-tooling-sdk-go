// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package xchain

// ChainType represents the X-Chain identifier
const ChainType = "X-Chain"

// TxType represents X-Chain transaction types
type TxType string

const (
	TransferTx TxType = "TransferTx"
	ExportTx   TxType = "ExportTx"
	ImportTx   TxType = "ImportTx"
)
