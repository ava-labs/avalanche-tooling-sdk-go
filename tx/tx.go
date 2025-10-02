// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package tx

import (
	"fmt"
	"math"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

type TxKind int64

var ErrUndefinedTx = fmt.Errorf("tx is undefined")

const (
	Undefined TxKind = iota
	PChainRemoveSubnetValidatorTx
	PChainAddSubnetValidatorTx
	PChainCreateChainTx
	PChainTransformSubnetTx
	PChainAddPermissionlessValidatorTx
	PChainTransferSubnetOwnershipTx
	PChainConvertSubnetToL1Tx
)

// BuildTxOutput represents a generic interface for transaction results
type BuildTxOutput interface {
	// GetTxType returns the transaction type identifier
	GetTxType() string
	// GetChainType returns which chain this transaction is for
	GetChainType() string
	// GetTx returns the actual transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
}

// PChainBuildTxResult represents a P-Chain transaction result
type PChainBuildTxResult struct {
	Tx *txs.Tx
}

func (p *PChainBuildTxResult) GetTxType() string {
	if p.Tx == nil || p.Tx.Unsigned == nil {
		return "Unknown"
	}
	// Extract tx type from unsigned transaction
	switch p.Tx.Unsigned.(type) {
	case *txs.CreateSubnetTx:
		return "CreateSubnetTx"
	case *txs.ConvertSubnetToL1Tx:
		return "ConvertSubnetToL1Tx"
	case *txs.AddSubnetValidatorTx:
		return "AddSubnetValidatorTx"
	case *txs.RemoveSubnetValidatorTx:
		return "RemoveSubnetValidatorTx"
	case *txs.CreateChainTx:
		return "CreateChainTx"
	case *txs.TransformSubnetTx:
		return "TransformSubnetTx"
	case *txs.AddPermissionlessValidatorTx:
		return "AddPermissionlessValidatorTx"
	case *txs.TransferSubnetOwnershipTx:
		return "TransferSubnetOwnershipTx"
	default:
		return "Unknown"
	}
}

func (p *PChainBuildTxResult) GetChainType() string {
	return "P-Chain"
}

func (p *PChainBuildTxResult) GetTx() interface{} {
	return p.Tx
}

func (p *PChainBuildTxResult) Validate() error {
	if p.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// CChainBuildTxResult represents a C-Chain transaction result
type CChainBuildTxResult struct {
	Tx interface{} // Will be *types.Transaction when C-Chain is implemented
}

func (c *CChainBuildTxResult) GetTxType() string {
	// TODO: Extract tx type from C-Chain transaction when implemented
	return "EVMTransaction"
}

func (c *CChainBuildTxResult) GetChainType() string {
	return "C-Chain"
}

func (c *CChainBuildTxResult) GetTx() interface{} {
	return c.Tx
}

func (c *CChainBuildTxResult) Validate() error {
	if c.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

// XChainBuildTxResult represents an X-Chain transaction result
type XChainBuildTxResult struct {
	Tx interface{} // Will be *avm.Tx when X-Chain is implemented
}

func (x *XChainBuildTxResult) GetTxType() string {
	// TODO: Extract tx type from X-Chain transaction when implemented
	return "AVMTransaction"
}

func (x *XChainBuildTxResult) GetChainType() string {
	return "X-Chain"
}

func (x *XChainBuildTxResult) GetTx() interface{} {
	return x.Tx
}

func (x *XChainBuildTxResult) Validate() error {
	if x.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

type BuildTxResult struct {
	BuildTxOutput
}

// Constructor functions for each chain type
func NewPChainBuildTxResult(tx *txs.Tx) *BuildTxResult {
	return &BuildTxResult{
		BuildTxOutput: &PChainBuildTxResult{Tx: tx},
	}
}

func NewCChainBuildTxResult(tx interface{}) *BuildTxResult {
	return &BuildTxResult{
		BuildTxOutput: &CChainBuildTxResult{Tx: tx},
	}
}

func NewXChainBuildTxResult(tx interface{}) *BuildTxResult {
	return &BuildTxResult{
		BuildTxOutput: &XChainBuildTxResult{Tx: tx},
	}
}

type SendTxResult struct {
	SignTxOutput
}

// SignTxOutput represents a generic interface for signed transaction results
type SignTxOutput interface {
	// GetTxType returns the transaction type identifier
	GetTxType() string
	// GetChainType returns which chain this transaction is for
	GetChainType() string
	// GetTx returns the actual signed transaction (interface{} to support different chain types)
	GetTx() interface{}
	// Validate validates the result
	Validate() error
	// IsReadyToCommit checks if the transaction is ready to be committed
	IsReadyToCommit() (bool, error)
}

// PChainSignTxResult represents a P-Chain signed transaction result
type PChainSignTxResult struct {
	Tx          *txs.Tx
	controlKeys []ids.ShortID
	threshold   uint32
}

func (p *PChainSignTxResult) GetTxType() string {
	if p.Tx == nil || p.Tx.Unsigned == nil {
		return "Unknown"
	}
	// Extract tx type from unsigned transaction
	switch p.Tx.Unsigned.(type) {
	case *txs.CreateSubnetTx:
		return "CreateSubnetTx"
	case *txs.ConvertSubnetToL1Tx:
		return "ConvertSubnetToL1Tx"
	case *txs.AddSubnetValidatorTx:
		return "AddSubnetValidatorTx"
	case *txs.RemoveSubnetValidatorTx:
		return "RemoveSubnetValidatorTx"
	case *txs.CreateChainTx:
		return "CreateChainTx"
	case *txs.TransformSubnetTx:
		return "TransformSubnetTx"
	case *txs.AddPermissionlessValidatorTx:
		return "AddPermissionlessValidatorTx"
	case *txs.TransferSubnetOwnershipTx:
		return "TransferSubnetOwnershipTx"
	default:
		return "Unknown"
	}
}

func (p *PChainSignTxResult) GetChainType() string {
	return "P-Chain"
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
		return false, ErrUndefinedTx
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

// CChainSignTxResult represents a C-Chain signed transaction result
type CChainSignTxResult struct {
	Tx interface{} // Will be *types.Transaction when C-Chain is implemented
}

func (c *CChainSignTxResult) GetTxType() string {
	// TODO: Extract tx type from C-Chain transaction when implemented
	return "EVMTransaction"
}

func (c *CChainSignTxResult) GetChainType() string {
	return "C-Chain"
}

func (c *CChainSignTxResult) GetTx() interface{} {
	return c.Tx
}

func (c *CChainSignTxResult) Validate() error {
	if c.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

func (c *CChainSignTxResult) IsReadyToCommit() (bool, error) {
	// TODO: Implement C-Chain ready to commit check when implemented
	return true, nil
}

// XChainSignTxResult represents an X-Chain signed transaction result
type XChainSignTxResult struct {
	Tx interface{} // Will be *avm.Tx when X-Chain is implemented
}

func (x *XChainSignTxResult) GetTxType() string {
	// TODO: Extract tx type from X-Chain transaction when implemented
	return "AVMTransaction"
}

func (x *XChainSignTxResult) GetChainType() string {
	return "X-Chain"
}

func (x *XChainSignTxResult) GetTx() interface{} {
	return x.Tx
}

func (x *XChainSignTxResult) Validate() error {
	if x.Tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	return nil
}

func (x *XChainSignTxResult) IsReadyToCommit() (bool, error) {
	// TODO: Implement X-Chain ready to commit check when implemented
	return true, nil
}

type SignTxResult struct {
	SignTxOutput
}

func (ms *SignTxResult) String() string {
	if ms.SignTxOutput != nil {
		if tx := ms.GetTx(); tx != nil {
			// For P-Chain transactions, we can get the ID
			if pChainTx, ok := tx.(*txs.Tx); ok {
				return pChainTx.ID().String()
			}
			// For other chains, return a generic representation
			return fmt.Sprintf("%s transaction", ms.GetChainType())
		}
	}
	return ""
}

func (ms *SignTxResult) Undefined() bool {
	return ms.SignTxOutput == nil || ms.GetTx() == nil
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
		return nil, nil, ErrUndefinedTx
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

func (ms *SignTxResult) GetRemainingAuthSigners() ([]ids.ShortID, []ids.ShortID, error) {
	if ms.Undefined() {
		return nil, nil, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetRemainingAuthSigners()
	}
	return nil, nil, fmt.Errorf("GetRemainingAuthSigners only supported for P-Chain transactions")
}

// GetAuthSigners gets all subnet auth addresses that are required to sign a given tx
//   - get subnet control keys as string slice using P-Chain API (GetOwners)
//   - get subnet auth indices from the tx, field tx.UnsignedTx.SubnetAuth
//   - creates the string slice of required subnet auth addresses by applying
//     the indices to the control keys slice
func (p *PChainSignTxResult) GetAuthSigners() ([]ids.ShortID, error) {
	if p.Tx == nil {
		return nil, ErrUndefinedTx
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

func (ms *SignTxResult) GetAuthSigners() ([]ids.ShortID, error) {
	if ms.Undefined() {
		return nil, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetAuthSigners()
	}
	return nil, fmt.Errorf("GetAuthSigners only supported for P-Chain transactions")
}

func (*SignTxResult) GetSpendSigners() ([]ids.ShortID, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (p *PChainSignTxResult) GetTxKind() (TxKind, error) {
	if p.Tx == nil {
		return Undefined, ErrUndefinedTx
	}
	unsignedTx := p.Tx.Unsigned
	switch unsignedTx := unsignedTx.(type) {
	case *txs.RemoveSubnetValidatorTx:
		return PChainRemoveSubnetValidatorTx, nil
	case *txs.AddSubnetValidatorTx:
		return PChainAddSubnetValidatorTx, nil
	case *txs.CreateChainTx:
		return PChainCreateChainTx, nil
	case *txs.TransformSubnetTx:
		return PChainTransformSubnetTx, nil
	case *txs.AddPermissionlessValidatorTx:
		return PChainAddPermissionlessValidatorTx, nil
	case *txs.TransferSubnetOwnershipTx:
		return PChainTransferSubnetOwnershipTx, nil
	case *txs.ConvertSubnetToL1Tx:
		return PChainConvertSubnetToL1Tx, nil
	default:
		return Undefined, fmt.Errorf("unable to GetTxKind due to unexpected unsigned tx type %T", unsignedTx)
	}
}

func (ms *SignTxResult) GetTxKind() (TxKind, error) {
	if ms.Undefined() {
		return Undefined, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetTxKind()
	}
	return Undefined, fmt.Errorf("GetTxKind only supported for P-Chain transactions")
}

// get network id associated to tx
func (p *PChainSignTxResult) GetNetworkID() (uint32, error) {
	if p.Tx == nil {
		return 0, ErrUndefinedTx
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
		return network.UndefinedNetwork, ErrUndefinedTx
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

func (ms *SignTxResult) GetNetworkID() (uint32, error) {
	if ms.Undefined() {
		return 0, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetNetworkID()
	}
	return 0, fmt.Errorf("GetNetworkID only supported for P-Chain transactions")
}

// get network model associated to tx
func (ms *SignTxResult) GetNetwork() (network.Network, error) {
	if ms.Undefined() {
		return network.UndefinedNetwork, ErrUndefinedTx
	}
	networkID, err := ms.GetNetworkID()
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
		return ids.Empty, ErrUndefinedTx
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

func (ms *SignTxResult) GetBlockchainID() (ids.ID, error) {
	if ms.Undefined() {
		return ids.Empty, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetBlockchainID()
	}
	return ids.Empty, fmt.Errorf("GetBlockchainID only supported for P-Chain transactions")
}

// GetSubnetID gets subnet id associated to tx
func (p *PChainSignTxResult) GetSubnetID() (ids.ID, error) {
	if p.Tx == nil {
		return ids.Empty, ErrUndefinedTx
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

func (ms *SignTxResult) GetSubnetID() (ids.ID, error) {
	if ms.Undefined() {
		return ids.Empty, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetSubnetID()
	}
	return ids.Empty, fmt.Errorf("GetSubnetID only supported for P-Chain transactions")
}

func (p *PChainSignTxResult) GetSubnetOwners() ([]ids.ShortID, uint32, error) {
	if p.Tx == nil {
		return nil, 0, ErrUndefinedTx
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

func (ms *SignTxResult) GetSubnetOwners() ([]ids.ShortID, uint32, error) {
	if ms.Undefined() {
		return nil, 0, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.GetSubnetOwners()
	}
	return nil, 0, fmt.Errorf("GetSubnetOwners only supported for P-Chain transactions")
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

func (ms *SignTxResult) GetWrappedPChainTx() (*txs.Tx, error) {
	if ms.Undefined() {
		return nil, ErrUndefinedTx
	}
	if pChainResult, ok := ms.SignTxOutput.(*PChainSignTxResult); ok {
		return pChainResult.Tx, nil
	}
	return nil, fmt.Errorf("GetWrappedPChainTx only supported for P-Chain transactions")
}

// Constructor functions for each chain type
func NewPChainSignTxResult(tx *txs.Tx) *SignTxResult {
	return &SignTxResult{
		SignTxOutput: &PChainSignTxResult{Tx: tx},
	}
}

func NewCChainSignTxResult(tx interface{}) *SignTxResult {
	return &SignTxResult{
		SignTxOutput: &CChainSignTxResult{Tx: tx},
	}
}

func NewXChainSignTxResult(tx interface{}) *SignTxResult {
	return &SignTxResult{
		SignTxOutput: &XChainSignTxResult{Tx: tx},
	}
}

// Constructor functions for SendTxResult
func NewPChainSendTxResult(tx *txs.Tx) *SendTxResult {
	return &SendTxResult{
		SignTxOutput: &PChainSignTxResult{Tx: tx},
	}
}

func NewCChainSendTxResult(tx interface{}) *SendTxResult {
	return &SendTxResult{
		SignTxOutput: &CChainSignTxResult{Tx: tx},
	}
}

func NewXChainSendTxResult(tx interface{}) *SendTxResult {
	return &SendTxResult{
		SignTxOutput: &XChainSignTxResult{Tx: tx},
	}
}
