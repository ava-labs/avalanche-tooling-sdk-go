// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"github.com/ava-labs/avalanchego/wallet/chain/c"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/subnet-evm/accounts/abi/bind"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
)

// Signer provides signing capabilities for EVM transactions
// It can be backed by a keychain (for real cryptographic signing) or used in NoOp mode
// (for generating unsigned transactions)
type Signer struct {
	signer keychain.Signer // nil when in NoOp mode
	addr   common.Address
}

// NewSigner creates a new EVM signer from an avalanchego EthKeychain
// This signer will perform actual cryptographic signing using the keychain
func NewSigner(kc c.EthKeychain) (*Signer, error) {
	if kc == nil {
		return nil, fmt.Errorf("keychain cannot be nil")
	}

	addrs := kc.EthAddresses()
	if len(addrs) != 1 {
		return nil, fmt.Errorf("expected keychain to have 1 address, found %d", len(addrs))
	}
	addr := addrs.List()[0]

	signer, ok := kc.GetEth(addr)
	if !ok {
		return nil, fmt.Errorf("unexpected failure obtaining unique signer from keychain")
	}

	return &Signer{
		signer: signer,
		addr:   addr,
	}, nil
}

// NewSignerFromPrivateKey creates a new EVM signer from a hex-encoded private key string
// This is a convenience function that handles the key loading and keychain creation
func NewSignerFromPrivateKey(privateKey string) (*Signer, error) {
	softKey, err := key.LoadSoftFromBytes([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return NewSigner(softKey.KeyChain())
}

// NewNoOpSigner creates a signer that doesn't actually sign transactions
// Useful for generating raw unsigned transactions where only the "from" address is needed
func NewNoOpSigner(addr common.Address) *Signer {
	return &Signer{
		addr: addr,
	}
}

// IsNoOp returns true if this signer is in NoOp mode (doesn't actually sign)
func (s *Signer) IsNoOp() bool {
	return s == nil || s.signer == nil
}

// SignTx signs the provided transaction with the given chainID and returns the signed transaction
// For NoOp signers, returns the transaction unchanged
func (s *Signer) SignTx(chainID *big.Int, tx *types.Transaction) (*types.Transaction, error) {
	if s == nil {
		return nil, fmt.Errorf("signer is nil")
	}
	if chainID == nil {
		return nil, fmt.Errorf("chainID cannot be nil")
	}
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	if s.IsNoOp() {
		return tx, nil
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
func (s *Signer) Address() common.Address {
	if s == nil {
		return common.Address{}
	}
	return s.addr
}

// TransactOpts returns a bind.TransactOpts configured for this signer
func (s *Signer) TransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
	if s == nil {
		return nil, fmt.Errorf("signer is nil")
	}
	if chainID == nil {
		return nil, fmt.Errorf("chainID cannot be nil")
	}

	opts := &bind.TransactOpts{
		From:    s.addr,
		Signer:  toSignerFn(s, chainID),
		Context: context.Background(),
	}

	if s.IsNoOp() {
		opts.NoSend = true
	}

	return opts, nil
}

// toSignerFn converts an EVM Signer into a bind.SignerFn that can be used with bind.TransactOpts
func toSignerFn(signer *Signer, chainID *big.Int) bind.SignerFn {
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != signer.Address() {
			return nil, fmt.Errorf("address mismatch: expected %s, got %s", signer.Address().Hex(), address.Hex())
		}
		return signer.SignTx(chainID, tx)
	}
}
