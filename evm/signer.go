// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/subnet-evm/accounts/abi/bind"
)

// Signer provides signing capabilities for EVM transactions
// It abstracts the underlying signing mechanism (private key, hardware wallet, avalanchego keychain, etc.)
type Signer interface {
	// SignTx signs the provided transaction with the given chainID and returns the signed transaction
	SignTx(chainID *big.Int, tx *types.Transaction) (*types.Transaction, error)
	// Address returns the EVM address associated with this signer
	Address() common.Address
	// TransactOpts returns a bind.TransactOpts configured for this signer
	TransactOpts(chainID *big.Int) (*bind.TransactOpts, error)
}

// AvalancheGoSigner wraps an avalanchego keychain.Signer to provide EVM signing capabilities
type AvalancheGoSigner struct {
	signer keychain.Signer
	addr   common.Address
}

// NewAvalancheGoSigner creates a new EVM signer from an avalanchego keychain signer and EVM address
func NewAvalancheGoSigner(signer keychain.Signer, addr common.Address) (*AvalancheGoSigner, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer cannot be nil")
	}

	return &AvalancheGoSigner{
		signer: signer,
		addr:   addr,
	}, nil
}

// SignTx implements the Signer interface by signing the transaction using the avalanchego signer
func (s *AvalancheGoSigner) SignTx(chainID *big.Int, tx *types.Transaction) (*types.Transaction, error) {
	if chainID == nil {
		return nil, fmt.Errorf("chainID cannot be nil")
	}
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	txSigner := types.LatestSignerForChainID(chainID)
	hash := txSigner.Hash(tx)

	signature, err := s.signer.SignHash(hash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction hash: %w", err)
	}

	return tx.WithSignature(txSigner, signature)
}

// Address returns the EVM address associated with this signer
func (s *AvalancheGoSigner) Address() common.Address {
	return s.addr
}

// TransactOpts returns a bind.TransactOpts configured for this signer
func (s *AvalancheGoSigner) TransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
	if chainID == nil {
		return nil, fmt.Errorf("chainID cannot be nil")
	}

	return &bind.TransactOpts{
		From:    s.addr,
		Signer:  toSignerFn(s, chainID),
		Context: context.Background(),
	}, nil
}

// NoOpSigner is a signer that doesn't actually sign transactions
// Useful for generating raw unsigned transactions where only the "from" address is needed
type NoOpSigner struct {
	addr common.Address
}

// NewNoOpSigner creates a new no-op signer with the given address
func NewNoOpSigner(addr common.Address) *NoOpSigner {
	return &NoOpSigner{
		addr: addr,
	}
}

// SignTx implements the Signer interface by returning the transaction unchanged (no actual signing)
func (*NoOpSigner) SignTx(_ *big.Int, tx *types.Transaction) (*types.Transaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}
	return tx, nil
}

// Address returns the EVM address associated with this signer
func (s *NoOpSigner) Address() common.Address {
	return s.addr
}

// TransactOpts returns a bind.TransactOpts configured for this signer
func (s *NoOpSigner) TransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
	if chainID == nil {
		return nil, fmt.Errorf("chainID cannot be nil")
	}

	return &bind.TransactOpts{
		From:    s.addr,
		Signer:  toSignerFn(s, chainID),
		Context: context.Background(),
		NoSend:  true,
	}, nil
}

// NullSigner is a sentinel value representing an absent/unset signer.
// Used to indicate that no signer was provided where a signer is optional.
type NullSigner struct{}

// NewNullSigner creates a new null signer
func NewNullSigner() *NullSigner {
	return &NullSigner{}
}

// SignTx implements the Signer interface but returns an error since null signer cannot sign
func (*NullSigner) SignTx(_ *big.Int, _ *types.Transaction) (*types.Transaction, error) {
	return nil, fmt.Errorf("null signer cannot sign transactions")
}

// Address returns the zero address
func (*NullSigner) Address() common.Address {
	return common.Address{}
}

// TransactOpts returns an error since null signer cannot create transaction options
func (*NullSigner) TransactOpts(_ *big.Int) (*bind.TransactOpts, error) {
	return nil, fmt.Errorf("null signer cannot create transaction options")
}

// toSignerFn converts an EVM Signer into a bind.SignerFn that can be used with bind.TransactOpts
func toSignerFn(signer Signer, chainID *big.Int) bind.SignerFn {
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != signer.Address() {
			return nil, fmt.Errorf("address mismatch: expected %s, got %s", signer.Address().Hex(), address.Hex())
		}
		return signer.SignTx(chainID, tx)
	}
}
