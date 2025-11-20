// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"
	"math"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	pchainTxs "github.com/ava-labs/avalanche-tooling-sdk-go/wallet/txs/p-chain"
)

// SignTxOutput represents a generic interface for signed transaction results
type SignTxOutput interface {
	// GetChainType returns which Avalanche chain this transaction is for.
	// Returns one of: "P-Chain", "X-Chain", "C-Chain"
	// Note: This is different from ChainID (blockchain identifier) or Network (Mainnet/Fuji/etc).
	GetChainType() string
	// GetTx returns the actual signed transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
	// IsReadyToCommit checks if the transaction is ready to be committed
	IsReadyToCommit() (bool, error)
}

// SignTxParams contains parameters for signing transactions
type SignTxParams struct {
	// AccountNames specifies which accounts to use for signing this transaction.
	// Currently only single-account transactions are supported (first element is used).
	// Future: Will support multi-account for multisig transactions.
	AccountNames []string
	*BuildTxResult
}

// Validate validates the sign transaction parameters
func (p *SignTxParams) Validate() error {
	// TODO: Support multiple accounts for multisig transactions
	if len(p.AccountNames) > 1 {
		return fmt.Errorf("only one account name is currently supported")
	}
	if p.BuildTxResult == nil {
		return fmt.Errorf("build tx result is required")
	}
	return p.BuildTxResult.Validate()
}

// SignTxResult represents the result of signing a transaction
type SignTxResult struct {
	SignTxOutput
}

// Validate validates the sign transaction result
func (r *SignTxResult) Validate() error {
	if r.SignTxOutput == nil {
		return fmt.Errorf("sign tx output is required")
	}
	return r.SignTxOutput.Validate()
}

// IsReadyToCommit checks if the transaction is ready to be committed
func (r *SignTxResult) IsReadyToCommit() (bool, error) {
	if r.SignTxOutput == nil {
		return false, fmt.Errorf("sign tx output is required")
	}
	return r.SignTxOutput.IsReadyToCommit()
}

// GetRemainingAuthSigners gets subnet auth addresses that have not signed a given tx
func (r *SignTxResult) GetRemainingAuthSigners() ([]ids.ShortID, []ids.ShortID, error) {
	if r.SignTxOutput == nil {
		return nil, nil, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetRemainingAuthSigners()
	}
	return nil, nil, fmt.Errorf("GetRemainingAuthSigners only supported for P-Chain transactions")
}

// GetAuthSigners gets all subnet auth addresses that are required to sign a given tx
func (r *SignTxResult) GetAuthSigners() ([]ids.ShortID, error) {
	if r.SignTxOutput == nil {
		return nil, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetAuthSigners()
	}
	return nil, fmt.Errorf("GetAuthSigners only supported for P-Chain transactions")
}

// GetTxKind returns the transaction kind
func (r *SignTxResult) GetTxKind() (pchainTxs.TxKind, error) {
	if r.SignTxOutput == nil {
		return pchainTxs.Undefined, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetTxKind()
	}
	return pchainTxs.Undefined, fmt.Errorf("GetTxKind only supported for P-Chain transactions")
}

// GetNetworkID returns the network ID associated with the transaction
func (r *SignTxResult) GetNetworkID() (uint32, error) {
	if r.SignTxOutput == nil {
		return 0, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetNetworkID()
	}
	return 0, fmt.Errorf("GetNetworkID only supported for P-Chain transactions")
}

// GetNetwork returns the network associated with the transaction
func (r *SignTxResult) GetNetwork() (network.Network, error) {
	if r.SignTxOutput == nil {
		return network.UndefinedNetwork, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetNetwork()
	}
	return network.UndefinedNetwork, fmt.Errorf("GetNetwork only supported for P-Chain transactions")
}

// GetBlockchainID returns the blockchain ID associated with the transaction
func (r *SignTxResult) GetBlockchainID() (ids.ID, error) {
	if r.SignTxOutput == nil {
		return ids.Empty, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetBlockchainID()
	}
	return ids.Empty, fmt.Errorf("GetBlockchainID only supported for P-Chain transactions")
}

// GetSubnetID returns the subnet ID associated with the transaction
func (r *SignTxResult) GetSubnetID() (ids.ID, error) {
	if r.SignTxOutput == nil {
		return ids.Empty, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetSubnetID()
	}
	return ids.Empty, fmt.Errorf("GetSubnetID only supported for P-Chain transactions")
}

// GetSubnetOwners returns the subnet owners
func (r *SignTxResult) GetSubnetOwners() ([]ids.ShortID, uint32, error) {
	if r.SignTxOutput == nil {
		return nil, 0, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetSubnetOwners()
	}
	return nil, 0, fmt.Errorf("GetSubnetOwners only supported for P-Chain transactions")
}

// GetWrappedPChainTx returns the wrapped P-Chain transaction
func (r *SignTxResult) GetWrappedPChainTx() (*txs.Tx, error) {
	if r.SignTxOutput == nil {
		return nil, fmt.Errorf("sign tx output is required")
	}
	// Delegate to the underlying SignTxOutput implementation
	if pChainResult, ok := r.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.Tx, nil
	}
	return nil, fmt.Errorf("GetWrappedPChainTx only supported for P-Chain transactions")
}

// String returns a string representation of the signed transaction
func (r *SignTxResult) String() string {
	if r.SignTxOutput != nil {
		if tx := r.GetTx(); tx != nil {
			// For P-Chain transactions, we can get the ID
			if pChainTx, ok := tx.(*txs.Tx); ok {
				return pChainTx.ID().String()
			}
			// For other chains, return a generic representation
			return fmt.Sprintf("%s transaction", r.GetChainType())
		}
	}
	return ""
}

// Undefined checks if the transaction is undefined
func (r *SignTxResult) Undefined() bool {
	return r.SignTxOutput == nil || r.GetTx() == nil
}

// PChainSignTxResult represents a P-Chain signed transaction result
type PChainSignTxResult struct {
	Tx          *txs.Tx
	controlKeys []ids.ShortID
	threshold   uint32
}

func (p *PChainSignTxResult) GetChainType() string {
	return constants.ChainTypePChain
}

func (p *PChainSignTxResult) GetTx() interface{} {
	return p.Tx
}

func (p *PChainSignTxResult) Validate() error {
	if p.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

func (p *PChainSignTxResult) IsReadyToCommit() (bool, error) {
	if p.Tx == nil {
		return false, pchainTxs.ErrUndefinedTx
	}
	unsignedTx := p.Tx.Unsigned
	switch unsignedTx.(type) {
	case *txs.CreateSubnetTx:
		return true, nil
	default:
	}
	_, remainingSigners, err := p.GetRemainingAuthSigners()
	if err != nil {
		return false, err
	}
	return len(remainingSigners) == 0, nil
}

// Constructor functions for each chain type
func NewPChainSignTxResult(tx *txs.Tx) *PChainSignTxResult {
	return &PChainSignTxResult{Tx: tx}
}

// GetRemainingAuthSigners gets subnet auth addresses that have not signed a given tx
//   - get the string slice of auth signers for the tx (GetAuthSigners)
//   - verifies that all creds in tx.Creds, except the last one, are fully signed
//     (a cred is fully signed if all the signatures in cred.Sigs are non-empty)
//   - computes remaining signers by iterating the last cred in tx.Creds, associated to subnet auth signing
//   - for each sig in cred.Sig: if sig is empty, then add the associated auth signer address (obtained from
//     authSigners by using the index) to the remaining signers list
//
// if the tx is fully signed, returns empty slice
func (p *PChainSignTxResult) GetRemainingAuthSigners() ([]ids.ShortID, []ids.ShortID, error) {
	if p.Tx == nil {
		return nil, nil, pchainTxs.ErrUndefinedTx
	}
	authSigners, err := p.GetAuthSigners()
	if err != nil {
		return nil, nil, err
	}
	emptySig := [secp256k1.SignatureLen]byte{}
	numCreds := len(p.Tx.Creds)
	// we should have at least 1 cred for output owners and 1 cred for subnet auth
	if numCreds < 2 {
		return nil, nil, fmt.Errorf("expected tx.Creds of len 2, got %d. doesn't seem to be a multisig tx with subnet auth requirements", numCreds)
	}
	// signatures for output owners should be filled (all creds except last one)
	for credIndex := range p.Tx.Creds[:numCreds-1] {
		cred, ok := p.Tx.Creds[credIndex].(*secp256k1fx.Credential)
		if !ok {
			return nil, nil, fmt.Errorf("expected cred to be of type *secp256k1fx.Credential, got %T", p.Tx.Creds[credIndex])
		}
		for i, sig := range cred.Sigs {
			if sig == emptySig {
				return nil, nil, fmt.Errorf("expected funding sig %d of cred %d to be filled", i, credIndex)
			}
		}
	}
	// signatures for subnet auth (last cred)
	cred, ok := p.Tx.Creds[numCreds-1].(*secp256k1fx.Credential)
	if !ok {
		return nil, nil, fmt.Errorf("expected cred to be of type *secp256k1fx.Credential, got %T", p.Tx.Creds[1])
	}
	if len(cred.Sigs) != len(authSigners) {
		return nil, nil, fmt.Errorf("expected number of cred's signatures %d to equal number of auth signers %d",
			len(cred.Sigs),
			len(authSigners),
		)
	}
	remainingSigners := []ids.ShortID{}
	for i, sig := range cred.Sigs {
		if sig == emptySig {
			remainingSigners = append(remainingSigners, authSigners[i])
		}
	}
	return authSigners, remainingSigners, nil
}

// GetAuthSigners gets all subnet auth addresses that are required to sign a given tx
//   - get subnet control keys as string slice using P-Chain API (GetOwners)
//   - get subnet auth indices from the tx, field tx.UnsignedTx.SubnetAuth
//   - creates the string slice of required subnet auth addresses by applying
//     the indices to the control keys slice
func (p *PChainSignTxResult) GetAuthSigners() ([]ids.ShortID, error) {
	if p.Tx == nil {
		return nil, pchainTxs.ErrUndefinedTx
	}
	controlKeys, _, err := p.GetSubnetOwners()
	if err != nil {
		return nil, err
	}
	unsignedTx := p.Tx.Unsigned
	var subnetAuth verify.Verifiable
	switch unsignedTx := unsignedTx.(type) {
	case *txs.RemoveSubnetValidatorTx:
		subnetAuth = unsignedTx.SubnetAuth
	case *txs.AddSubnetValidatorTx:
		subnetAuth = unsignedTx.SubnetAuth
	case *txs.CreateChainTx:
		subnetAuth = unsignedTx.SubnetAuth
	case *txs.TransformSubnetTx:
		subnetAuth = unsignedTx.SubnetAuth
	case *txs.TransferSubnetOwnershipTx:
		subnetAuth = unsignedTx.SubnetAuth
	case *txs.ConvertSubnetToL1Tx:
		subnetAuth = unsignedTx.SubnetAuth
	default:
		return nil, fmt.Errorf("unable to GetAuthSigners due to unexpected unsigned tx type %T", unsignedTx)
	}
	subnetInput, ok := subnetAuth.(*secp256k1fx.Input)
	if !ok {
		return nil, fmt.Errorf("expected subnetAuth of type *secp256k1fx.Input, got %T", subnetAuth)
	}
	controlKeysLen := len(controlKeys)
	if controlKeysLen > math.MaxUint32 {
		return nil, fmt.Errorf("value %d out of range for uint32", controlKeysLen)
	}
	authSigners := []ids.ShortID{}
	for _, sigIndex := range subnetInput.SigIndices {
		if sigIndex >= uint32(controlKeysLen) {
			return nil, fmt.Errorf("signer index %d exceeds number of control keys", sigIndex)
		}
		authSigners = append(authSigners, controlKeys[sigIndex])
	}
	return authSigners, nil
}

func (p *PChainSignTxResult) GetTxKind() (pchainTxs.TxKind, error) {
	if p.Tx == nil {
		return pchainTxs.Undefined, pchainTxs.ErrUndefinedTx
	}
	unsignedTx := p.Tx.Unsigned
	switch unsignedTx := unsignedTx.(type) {
	case *txs.RemoveSubnetValidatorTx:
		return pchainTxs.PChainRemoveSubnetValidatorTx, nil
	case *txs.AddSubnetValidatorTx:
		return pchainTxs.PChainAddSubnetValidatorTx, nil
	case *txs.CreateChainTx:
		return pchainTxs.PChainCreateChainTx, nil
	case *txs.TransformSubnetTx:
		return pchainTxs.PChainTransformSubnetTx, nil
	case *txs.AddPermissionlessValidatorTx:
		return pchainTxs.PChainAddPermissionlessValidatorTx, nil
	case *txs.TransferSubnetOwnershipTx:
		return pchainTxs.PChainTransferSubnetOwnershipTx, nil
	case *txs.ConvertSubnetToL1Tx:
		return pchainTxs.PChainConvertSubnetToL1Tx, nil
	default:
		return pchainTxs.Undefined, fmt.Errorf("unable to GetTxKind due to unexpected unsigned tx type %T", unsignedTx)
	}
}

// get network id associated to tx
func (p *PChainSignTxResult) GetNetworkID() (uint32, error) {
	if p.Tx == nil {
		return 0, pchainTxs.ErrUndefinedTx
	}
	unsignedTx := p.Tx.Unsigned
	var networkID uint32
	switch unsignedTx := unsignedTx.(type) {
	case *txs.RemoveSubnetValidatorTx:
		networkID = unsignedTx.NetworkID
	case *txs.AddSubnetValidatorTx:
		networkID = unsignedTx.NetworkID
	case *txs.CreateChainTx:
		networkID = unsignedTx.NetworkID
	case *txs.TransformSubnetTx:
		networkID = unsignedTx.NetworkID
	case *txs.AddPermissionlessValidatorTx:
		networkID = unsignedTx.NetworkID
	case *txs.TransferSubnetOwnershipTx:
		networkID = unsignedTx.NetworkID
	case *txs.ConvertSubnetToL1Tx:
		networkID = unsignedTx.NetworkID
	default:
		return 0, fmt.Errorf("unable to GetNetworkID due to unexpected unsigned tx type %T", unsignedTx)
	}
	return networkID, nil
}

func (p *PChainSignTxResult) GetNetwork() (network.Network, error) {
	if p.Tx == nil {
		return network.UndefinedNetwork, pchainTxs.ErrUndefinedTx
	}
	networkID, err := p.GetNetworkID()
	if err != nil {
		return network.UndefinedNetwork, err
	}
	newNetwork := network.NetworkFromNetworkID(networkID)
	if newNetwork.Kind == network.Undefined {
		return network.UndefinedNetwork, fmt.Errorf("undefined network model for tx")
	}
	return newNetwork, nil
}

func (p *PChainSignTxResult) GetBlockchainID() (ids.ID, error) {
	if p.Tx == nil {
		return ids.Empty, pchainTxs.ErrUndefinedTx
	}
	unsignedTx := p.Tx.Unsigned
	var blockchainID ids.ID
	switch unsignedTx := unsignedTx.(type) {
	case *txs.RemoveSubnetValidatorTx:
		blockchainID = unsignedTx.BlockchainID
	case *txs.AddSubnetValidatorTx:
		blockchainID = unsignedTx.BlockchainID
	case *txs.CreateChainTx:
		blockchainID = unsignedTx.BlockchainID
	case *txs.TransformSubnetTx:
		blockchainID = unsignedTx.BlockchainID
	case *txs.AddPermissionlessValidatorTx:
		blockchainID = unsignedTx.BlockchainID
	case *txs.TransferSubnetOwnershipTx:
		blockchainID = unsignedTx.BlockchainID
	case *txs.ConvertSubnetToL1Tx:
		blockchainID = unsignedTx.BlockchainID
	default:
		return ids.Empty, fmt.Errorf("unable to GetBlockchainID due to unexpected unsigned tx type %T", unsignedTx)
	}
	return blockchainID, nil
}

// GetSubnetID gets subnet id associated to tx
func (p *PChainSignTxResult) GetSubnetID() (ids.ID, error) {
	if p.Tx == nil {
		return ids.Empty, pchainTxs.ErrUndefinedTx
	}
	unsignedTx := p.Tx.Unsigned
	var subnetID ids.ID
	switch unsignedTx := unsignedTx.(type) {
	case *txs.RemoveSubnetValidatorTx:
		subnetID = unsignedTx.Subnet
	case *txs.AddSubnetValidatorTx:
		subnetID = unsignedTx.SubnetValidator.Subnet
	case *txs.CreateChainTx:
		subnetID = unsignedTx.SubnetID
	case *txs.TransformSubnetTx:
		subnetID = unsignedTx.Subnet
	case *txs.AddPermissionlessValidatorTx:
		subnetID = unsignedTx.Subnet
	case *txs.TransferSubnetOwnershipTx:
		subnetID = unsignedTx.Subnet
	case *txs.ConvertSubnetToL1Tx:
		subnetID = unsignedTx.Subnet
	default:
		return ids.Empty, fmt.Errorf("unable to GetSubnetID due to unexpected unsigned tx type %T", unsignedTx)
	}
	return subnetID, nil
}

func (p *PChainSignTxResult) GetSubnetOwners() ([]ids.ShortID, uint32, error) {
	if p.Tx == nil {
		return nil, 0, pchainTxs.ErrUndefinedTx
	}
	if p.controlKeys == nil {
		subnetID, err := p.GetSubnetID()
		if err != nil {
			return nil, 0, err
		}

		network, err := p.GetNetwork()
		if err != nil {
			return nil, 0, err
		}
		controlKeys, threshold, err := GetOwners(network, subnetID)
		if err != nil {
			return nil, 0, err
		}
		p.controlKeys = controlKeys
		p.threshold = threshold
	}
	return p.controlKeys, p.threshold, nil
}

func GetOwners(network network.Network, subnetID ids.ID) ([]ids.ShortID, uint32, error) {
	pClient := platformvm.NewClient(network.Endpoint)
	ctx, cancel := utils.GetAPIContext()
	defer cancel()
	subnetResponse, err := pClient.GetSubnet(ctx, subnetID)
	if err != nil {
		return nil, 0, fmt.Errorf("subnet tx %s query error: %w", subnetID, err)
	}
	controlKeys := subnetResponse.ControlKeys
	threshold := subnetResponse.Threshold
	return controlKeys, threshold, nil
}
