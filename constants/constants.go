// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package constants

import "time"

const (
	// http
	APIRequestTimeout      = 10 * time.Second
	APIRequestLargeTimeout = 30 * time.Second

	// node
	UserOnlyWriteReadPerms     = 0o600
	UserOnlyWriteReadExecPerms = 0o700
	WriteReadUserOnlyPerms     = 0o600

	SignatureTimeout = 5 * time.Minute
)

// Transaction type constants for all chains

// P-Chain transaction types
const (
	PChainCreateSubnetTx               = "CreateSubnetTx"
	PChainConvertSubnetToL1Tx          = "ConvertSubnetToL1Tx"
	PChainRegisterL1ValidatorTx        = "RegisterL1ValidatorTx"
	PChainAddSubnetValidatorTx         = "AddSubnetValidatorTx"
	PChainRemoveSubnetValidatorTx      = "RemoveSubnetValidatorTx"
	PChainCreateChainTx                = "CreateChainTx"
	PChainTransformSubnetTx            = "TransformSubnetTx"
	PChainAddPermissionlessValidatorTx = "AddPermissionlessValidatorTx"
	PChainTransferSubnetOwnershipTx    = "TransferSubnetOwnershipTx"
)

// C-Chain transaction types
const (
	CChainTransferTx     = "TransferTx"
	CChainContractCallTx = "ContractCallTx"
	CChainDeployTx       = "DeployTx"
)

// X-Chain transaction types
const (
	XChainTransferTx = "TransferTx"
	XChainExportTx   = "ExportTx"
	XChainImportTx   = "ImportTx"
)

// Chain type constants
const (
	ChainTypePChain = "P-Chain"
	ChainTypeCChain = "C-Chain"
	ChainTypeXChain = "X-Chain"
)

// Transaction type constants
const (
	TxTypeUnknown        = "Unknown"
	TxTypeEVMTransaction = "EVMTransaction"
	TxTypeAVMTransaction = "AVMTransaction"
)
