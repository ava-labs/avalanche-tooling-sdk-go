// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import "fmt"

// TxKind represents the type of P-Chain transaction
type TxKind int64

// ErrUndefinedTx is returned when a transaction is undefined
var ErrUndefinedTx = fmt.Errorf("tx is undefined")

const (
	Undefined TxKind = iota
	PChainRemoveSubnetValidatorTx
	PChainAddSubnetValidatorTx
	PChainCreateChainTx
	PChainTransformSubnetTx
	PChainAddPermissionlessValidatorTx
	PChainTransferSubnetOwnershipTx
	PChainConvertSubnetToL1Tx
	PChainRegisterL1ValidatorTx
)
