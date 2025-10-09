// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// ConvertSubnetToL1TxParams contains all parameters needed to create a ConvertSubnetToL1Tx
type ConvertSubnetToL1Validator struct {
	// NodeID of this validator
	NodeID string `serialize:"true" json:"nodeID"`
	// Weight of this validator used when sampling
	Weight uint64 `serialize:"true" json:"weight"`
	// Initial balance for this validator
	Balance uint64 `serialize:"true" json:"balance"`
	// [Signer] is the BLS key for this validator.
	// Note: We do not enforce that the BLS key is unique across all validators.
	//       This means that validators can share a key if they so choose.
	//       However, a NodeID + Subnet does uniquely map to a BLS key
	BLSPublicKey         string `serialize:"true" json:"signer"`
	BLSProofOfPossession string
	// Leftover $AVAX from the [Balance] will be issued to this owner once it is
	// removed from the validator set.
	RemainingBalanceOwner string `serialize:"true" json:"remainingBalanceOwner"`
	// This owner has the authority to manually deactivate this validator.
	DeactivationOwner string `serialize:"true" json:"deactivationOwner"`
}

// ConvertSubnetToL1TxParams contains all parameters needed to create a ConvertSubnetToL1Tx
type ConvertSubnetToL1TxParams struct {
	// SubnetAuthKeys are the keys used to sign `ConvertSubnetToL1Tx`
	SubnetAuthKeys []string
	// SubnetID is Subnet ID of the subnet to convert to an L1.
	SubnetID string
	// ChainID is Blockchain ID of the L1 where the validator manager contract is deployed.
	ChainID string
	// Address is address of the validator manager contract.
	Address string
	// Validators are the initial set of L1 validators after the conversion.
	Validators []*ConvertSubnetToL1Validator
}

// GetTxType returns the transaction type identifier
func (p ConvertSubnetToL1TxParams) GetTxType() string {
	return constants.PChainConvertSubnetToL1Tx
}

// Validate validates the parameters
func (p ConvertSubnetToL1TxParams) Validate() error {
	if len(p.SubnetAuthKeys) == 0 {
		return fmt.Errorf("subnet auth keys cannot be empty")
	}
	if p.SubnetID == "" {
		return fmt.Errorf("subnet ID cannot be empty")
	}
	if p.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if len(p.Address) == 0 {
		return fmt.Errorf("address cannot be empty")
	}
	if len(p.Validators) == 0 {
		return fmt.Errorf("validators cannot be empty")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p ConvertSubnetToL1TxParams) GetChainType() string {
	return constants.ChainTypePChain
}
