// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package pchain

// ChainType represents the P-Chain identifier
const ChainType = "P-Chain"

// TxType represents P-Chain transaction types
type TxType string

const (
	CreateSubnetTx               TxType = "CreateSubnetTx"
	ConvertSubnetToL1Tx          TxType = "ConvertSubnetToL1Tx"
	AddSubnetValidatorTx         TxType = "AddSubnetValidatorTx"
	RemoveSubnetValidatorTx      TxType = "RemoveSubnetValidatorTx"
	CreateChainTx                TxType = "CreateChainTx"
	TransformSubnetTx            TxType = "TransformSubnetTx"
	AddPermissionlessValidatorTx TxType = "AddPermissionlessValidatorTx"
	TransferSubnetOwnershipTx    TxType = "TransferSubnetOwnershipTx"
)
