// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

// Transaction type constants for all chains

// P-Chain transaction types
const (
	PChainCreateSubnetTx               = "CreateSubnetTx"
	PChainConvertSubnetToL1Tx          = "ConvertSubnetToL1Tx"
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
