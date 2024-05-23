// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package multisig

import (
	"avalanche-tooling-sdk-go/avalanche"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

type TxKind struct {
	_ string // vm
	_ string // tx
}

type Multisig struct {
	_ *txs.Tx // pChainTx
}

func New(_ *txs.Tx) *Multisig {
	return nil
}

func (*Multisig) ToBytes() ([]byte, error) {
	return nil, nil
}

func (*Multisig) FromBytes(_ []byte) error {
	return nil
}

func (*Multisig) ToFile(_ string) error {
	return nil
}

func (*Multisig) FromFile(_ string) error {
	return nil
}

func (*Multisig) Sign(_ *primary.Wallet) error {
	return nil
}

func (*Multisig) Commit() error {
	return nil
}

func (*Multisig) IsReadyToCommit() error {
	return nil
}

func (*Multisig) GetRemainingSigners() ([]ids.ID, error) {
	return nil, nil
}

func (*Multisig) GetAuthSigners() ([]ids.ID, error) {
	return nil, nil
}

func (*Multisig) GetFeeSigners() ([]ids.ID, error) {
	return nil, nil
}

func (*Multisig) GetTxKind() TxKind {
	return TxKind{}
}

func (*Multisig) GetNetwork() (avalanche.Network, error) {
	return avalanche.Network{}, nil
}

func (*Multisig) GetBlockchainID() (ids.ID, error) {
	return ids.Empty, nil
}

func (*Multisig) GetSubnetID() (ids.ID, error) {
	return ids.Empty, nil
}

func (*Multisig) GetSubnetOwners() ([]ids.ID, int, error) {
	return nil, 0, nil
}
