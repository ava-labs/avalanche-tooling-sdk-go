// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package cchain

// ChainType represents the C-Chain identifier
const ChainType = "C-Chain"

// TxType represents C-Chain transaction types
type TxType string

const (
	TransferTx     TxType = "TransferTx"
	ContractCallTx TxType = "ContractCallTx"
	DeployTx       TxType = "DeployTx"
)
