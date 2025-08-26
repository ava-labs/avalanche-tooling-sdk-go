// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package transaction

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	ptxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ethereum/go-ethereum/core/types"
)

// ChainType represents the different blockchain types in Avalanche
type ChainType int

const (
	PChain ChainType = iota
	CChain
)

func (ct ChainType) String() string {
	switch ct {
	case PChain:
		return "P-Chain"
	case CChain:
		return "C-Chain"
	default:
		return "Unknown"
	}
}

// Transaction represents a generic transaction interface that can handle
// P-Chain and C-Chain transactions in Avalanche
type Transaction interface {
	// GetChainType returns the chain type this transaction belongs to
	GetChainType() ChainType

	// GetID returns the transaction ID
	GetID() (ids.ID, error)

	// GetHash returns the transaction hash (for C-Chain compatibility)
	GetHash() (string, error)

	// IsSigned returns whether the transaction has been signed
	IsSigned() bool

	// GetBytes returns the serialized transaction bytes
	GetBytes() ([]byte, error)

	// GetNetworkID returns the network ID associated with the transaction
	GetNetworkID() (uint32, error)
}

// PChainTransaction wraps a P-Chain transaction
type PChainTransaction struct {
	Tx *ptxs.Tx
}

// NewPChainTransaction creates a new P-Chain transaction wrapper
func NewPChainTransaction(tx *ptxs.Tx) *PChainTransaction {
	return &PChainTransaction{Tx: tx}
}

func (p *PChainTransaction) GetChainType() ChainType {
	return PChain
}

func (p *PChainTransaction) GetID() (ids.ID, error) {
	if p.Tx == nil {
		return ids.Empty, fmt.Errorf("P-Chain transaction is nil")
	}
	return p.Tx.ID(), nil
}

func (p *PChainTransaction) GetHash() (string, error) {
	id, err := p.GetID()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (p *PChainTransaction) IsSigned() bool {
	return p.Tx != nil && len(p.Tx.Creds) > 0
}

func (p *PChainTransaction) GetBytes() ([]byte, error) {
	if p.Tx == nil {
		return nil, fmt.Errorf("P-Chain transaction is nil")
	}
	return ptxs.Codec.Marshal(ptxs.CodecVersion, p.Tx)
}

func (p *PChainTransaction) GetNetworkID() (uint32, error) {
	if p.Tx == nil || p.Tx.Unsigned == nil {
		return 0, fmt.Errorf("P-Chain transaction or unsigned transaction is nil")
	}

	// Extract network ID from the unsigned transaction
	switch unsignedTx := p.Tx.Unsigned.(type) {
	case *ptxs.RemoveSubnetValidatorTx:
		return unsignedTx.NetworkID, nil
	case *ptxs.AddSubnetValidatorTx:
		return unsignedTx.NetworkID, nil
	case *ptxs.CreateChainTx:
		return unsignedTx.NetworkID, nil
	case *ptxs.TransformSubnetTx:
		return unsignedTx.NetworkID, nil
	case *ptxs.AddPermissionlessValidatorTx:
		return unsignedTx.NetworkID, nil
	case *ptxs.TransferSubnetOwnershipTx:
		return unsignedTx.NetworkID, nil
	case *ptxs.ConvertSubnetToL1Tx:
		return unsignedTx.NetworkID, nil
	case *ptxs.CreateSubnetTx:
		return unsignedTx.NetworkID, nil
	default:
		return 0, fmt.Errorf("unsupported P-Chain transaction type: %T", unsignedTx)
	}
}

// CChainTransaction wraps a C-Chain (Ethereum) transaction
type CChainTransaction struct {
	Tx *types.Transaction
}

// NewCChainTransaction creates a new C-Chain transaction wrapper
func NewCChainTransaction(tx *types.Transaction) *CChainTransaction {
	return &CChainTransaction{Tx: tx}
}

func (c *CChainTransaction) GetChainType() ChainType {
	return CChain
}

func (c *CChainTransaction) GetID() (ids.ID, error) {
	if c.Tx == nil {
		return ids.Empty, fmt.Errorf("C-Chain transaction is nil")
	}
	// For C-Chain, we convert the hash to an ID
	hash := c.Tx.Hash()
	id, err := ids.ToID(hash[:])
	if err != nil {
		return ids.Empty, fmt.Errorf("failed to convert hash to ID: %w", err)
	}
	return id, nil
}

func (c *CChainTransaction) GetHash() (string, error) {
	if c.Tx == nil {
		return "", fmt.Errorf("C-Chain transaction is nil")
	}
	return c.Tx.Hash().Hex(), nil
}

func (c *CChainTransaction) IsSigned() bool {
	// For C-Chain, we assume the transaction is signed if it exists
	// since Ethereum transactions are typically signed when created
	return c.Tx != nil
}

func (c *CChainTransaction) GetBytes() ([]byte, error) {
	if c.Tx == nil {
		return nil, fmt.Errorf("C-Chain transaction is nil")
	}
	return c.Tx.MarshalBinary()
}

func (c *CChainTransaction) GetNetworkID() (uint32, error) {
	if c.Tx == nil {
		return 0, fmt.Errorf("C-Chain transaction is nil")
	}
	// For C-Chain, we extract the chain ID from the transaction
	chainID := c.Tx.ChainId()
	if chainID == nil {
		return 0, fmt.Errorf("C-Chain transaction has no chain ID")
	}
	return uint32(chainID.Uint64()), nil
}

// TransactionSigner provides methods to sign transactions for different chains
type TransactionSigner interface {
	// SignTransaction signs a transaction using the provided key
	SignTransaction(tx Transaction, key interface{}) error

	// SignPChainTransaction signs a P-Chain transaction
	SignPChainTransaction(tx *PChainTransaction, key interface{}) error

	// SignCChainTransaction signs a C-Chain transaction
	SignCChainTransaction(tx *CChainTransaction, key interface{}) error
}

// SignTx is a generic function that can sign any type of transaction
// and returns a generic Transaction interface
func SignTx(tx interface{}, key interface{}) (Transaction, error) {
	switch t := tx.(type) {
	case *ptxs.Tx:
		// Handle P-Chain transaction
		pChainTx := NewPChainTransaction(t)
		if err := SignPChainTransaction(pChainTx, key); err != nil {
			return nil, fmt.Errorf("failed to sign P-Chain transaction: %w", err)
		}
		return pChainTx, nil

	case *types.Transaction:
		// Handle C-Chain transaction
		cChainTx := NewCChainTransaction(t)
		if err := SignCChainTransaction(cChainTx, key); err != nil {
			return nil, fmt.Errorf("failed to sign C-Chain transaction: %w", err)
		}
		return cChainTx, nil

	default:
		return nil, fmt.Errorf("unsupported transaction type: %T", tx)
	}
}

// SignPChainTransaction signs a P-Chain transaction
func SignPChainTransaction(tx *PChainTransaction, key interface{}) error {
	if tx == nil || tx.Tx == nil {
		return fmt.Errorf("P-Chain transaction is nil")
	}

	// This is a placeholder implementation
	// In a real implementation, you would use the key to sign the transaction
	// For now, we'll just mark it as signed by ensuring it has credentials
	if len(tx.Tx.Creds) == 0 {
		return fmt.Errorf("P-Chain transaction signing not implemented yet")
	}

	return nil
}

// SignCChainTransaction signs a C-Chain transaction
func SignCChainTransaction(tx *CChainTransaction, key interface{}) error {
	if tx == nil || tx.Tx == nil {
		return fmt.Errorf("C-Chain transaction is nil")
	}

	// This is a placeholder implementation
	// In a real implementation, you would use the key to sign the transaction
	// For now, we'll just return success since Ethereum transactions are typically pre-signed
	return nil
}

// Utility functions for working with transactions

// GetTransactionType returns the chain type of a transaction
func GetTransactionType(tx Transaction) ChainType {
	return tx.GetChainType()
}

// IsPChainTransaction checks if a transaction is a P-Chain transaction
func IsPChainTransaction(tx Transaction) bool {
	return tx.GetChainType() == PChain
}

// IsCChainTransaction checks if a transaction is a C-Chain transaction
func IsCChainTransaction(tx Transaction) bool {
	return tx.GetChainType() == CChain
}

// GetTransactionID returns the transaction ID as a string
func GetTransactionID(tx Transaction) (string, error) {
	id, err := tx.GetID()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// GetTransactionHash returns the transaction hash
func GetTransactionHash(tx Transaction) (string, error) {
	return tx.GetHash()
}

// IsTransactionSigned checks if a transaction is signed
func IsTransactionSigned(tx Transaction) bool {
	return tx.IsSigned()
}

// CreateTransactionFromRaw creates a generic transaction from raw transaction data
func CreateTransactionFromRaw(tx interface{}) (Transaction, error) {
	switch t := tx.(type) {
	case *ptxs.Tx:
		return NewPChainTransaction(t), nil
	case *types.Transaction:
		return NewCChainTransaction(t), nil
	default:
		return nil, fmt.Errorf("unsupported transaction type: %T", tx)
	}
}
