// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package xchain

import (
	"errors"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x/builder"
)

// GetNetworkID extracts the NetworkID from an UnsignedTx.
// It performs type assertion on all known X-Chain transaction types and returns the NetworkID.
// Returns an error if the transaction type is not supported or recognized.
func GetNetworkID(utx txs.UnsignedTx) (uint32, error) {
	switch tx := utx.(type) {
	case *txs.BaseTx:
		return tx.NetworkID, nil
	case *txs.CreateAssetTx:
		return tx.NetworkID, nil
	case *txs.OperationTx:
		return tx.NetworkID, nil
	case *txs.ImportTx:
		return tx.NetworkID, nil
	case *txs.ExportTx:
		return tx.NetworkID, nil
	default:
		return 0, errors.New("unsupported transaction type: unable to extract NetworkID")
	}
}

// TxFromBytes attempts to unmarshal bytes into an X-Chain unsigned transaction.
// Returns the transaction and true if successful, nil and false otherwise.
func TxFromBytes(b []byte) (txs.UnsignedTx, bool) {
	if len(b) == 0 {
		return nil, false
	}

	tx := new(txs.UnsignedTx)
	if _, err := builder.Parser.Codec().Unmarshal(b, tx); err != nil {
		return nil, false
	}

	return *tx, true
}

// IsXChainTx checks if the provided bytes represent an X-Chain unsigned transaction.
func IsXChainTx(b []byte) bool {
	_, ok := TxFromBytes(b)
	return ok
}

// GetHRP extracts the HRP (Human Readable Part) from an X-Chain unsigned transaction.
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
