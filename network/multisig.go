// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package multisig

import (
	"avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

type PChainTxKind int

const (
	Invalid = iota
	CreateBlockchain
	TransferSubnetOwnership
)

type PChainMultisig struct {
	_ *txs.Tx
}

func New(_ *txs.Tx) *PChainMultisig {
	return nil
}

func (*PChainMultisig) ToBytes() ([]byte, error) {
	return nil, nil
}

func (*PChainMultisig) FromBytes(_ []byte) error {
	return nil
}

func (*PChainMultisig) ToFile(_ string) error {
	return nil
}

func (*PChainMultisig) FromFile(_ string) error {
	return nil
}

func (*PChainMultisig) Sign(_ *primary.Wallet) error {
	return nil
}

func (*PChainMultisig) Commit() error {
	return nil
}

func (*PChainMultisig) IsReadyToCommit() error {
	return nil
}

func (*PChainMultisig) GetRemainingSigners() ([]ids.ID, error) {
	return nil, nil
}

func (*PChainMultisig) GetAuthSigners() ([]ids.ID, error) {
	return nil, nil
}

func (*PChainMultisig) GetFeeSigners() ([]ids.ID, error) {
	return nil, nil
}

func (*PChainMultisig) GetKind() PChainTxKind {
	return Invalid
}

func (*PChainMultisig) GetNetwork() (avalanche.Network, error) {
	return avalanche.UndefinedNetwork, nil
}

func (*PChainMultisig) GetSubnetID() (ids.ID, error) {
	return ids.Empty, nil
}

func (*PChainMultisig) GetSubnetOwners() ([]ids.ID, int, error) {
	return nil, 0, nil
}
