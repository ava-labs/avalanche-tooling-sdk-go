// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package pchain

import (
	"errors"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// GetNetworkID extracts the NetworkID from an UnsignedTx.
// It performs type assertion on all known P-Chain transaction types and returns the NetworkID.
// Returns an error if the transaction type is not supported or recognized.
func GetNetworkID(utx txs.UnsignedTx) (uint32, error) {
	switch tx := utx.(type) {
	case *txs.AddDelegatorTx:
		return tx.NetworkID, nil
	case *txs.AddPermissionlessDelegatorTx:
		return tx.NetworkID, nil
	case *txs.AddPermissionlessValidatorTx:
		return tx.NetworkID, nil
	case *txs.AddSubnetValidatorTx:
		return tx.NetworkID, nil
	case *txs.AddValidatorTx:
		return tx.NetworkID, nil
	case *txs.BaseTx:
		return tx.NetworkID, nil
	case *txs.ConvertSubnetToL1Tx:
		return tx.NetworkID, nil
	case *txs.CreateChainTx:
		return tx.NetworkID, nil
	case *txs.CreateSubnetTx:
		return tx.NetworkID, nil
	case *txs.DisableL1ValidatorTx:
		return tx.NetworkID, nil
	case *txs.ExportTx:
		return tx.NetworkID, nil
	case *txs.ImportTx:
		return tx.NetworkID, nil
	case *txs.IncreaseL1ValidatorBalanceTx:
		return tx.NetworkID, nil
	case *txs.RegisterL1ValidatorTx:
		return tx.NetworkID, nil
	case *txs.RemoveSubnetValidatorTx:
		return tx.NetworkID, nil
	case *txs.SetL1ValidatorWeightTx:
		return tx.NetworkID, nil
	case *txs.TransferSubnetOwnershipTx:
		return tx.NetworkID, nil
	case *txs.TransformSubnetTx:
		return tx.NetworkID, nil
	default:
		return 0, errors.New("unsupported transaction type: unable to extract NetworkID")
	}
}

// TxFromBytes attempts to unmarshal bytes into a P-Chain unsigned transaction.
// Returns the transaction and true if successful, nil and false otherwise.
func TxFromBytes(b []byte) (txs.UnsignedTx, bool) {
	if len(b) == 0 {
		return nil, false
	}

	var unsignedTx txs.UnsignedTx
	if _, err := txs.Codec.Unmarshal(b, &unsignedTx); err != nil {
		return nil, false
	}

	return unsignedTx, true
}

// IsPChainTx checks if the provided bytes represent a P-Chain unsigned transaction.
func IsPChainTx(b []byte) bool {
	_, ok := TxFromBytes(b)
	return ok
}

// GetHRP extracts the HRP (Human Readable Part) from a P-Chain unsigned transaction.
// It first extracts the NetworkID and then maps it to the corresponding HRP using
// the centralized network.HRPFromNetworkID function.
// Returns an error if the transaction type is not supported.
func GetHRP(utx txs.UnsignedTx) (string, error) {
	networkID, err := GetNetworkID(utx)
	if err != nil {
		return "", err
	}

	return network.HRPFromNetworkID(networkID), nil
}
