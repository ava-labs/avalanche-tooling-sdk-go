// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

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
	Validators []*avagoTxs.ConvertSubnetToL1Validator
}

// GetTxType returns the transaction type identifier
func (p ConvertSubnetToL1TxParams) GetTxType() string {
	return "ConvertSubnetToL1Tx"
}

// Validate validates the parameters
func (p ConvertSubnetToL1TxParams) Validate() error {
	if len(p.SubnetAuthKeys) == 0 {
		return fmt.Errorf("SubnetAuthKeys cannot be empty")
	}
	if p.SubnetID == "" {
		return fmt.Errorf("SubnetID cannot be empty")
	}
	if p.ChainID == "" {
		return fmt.Errorf("ChainID cannot be empty")
	}
	if len(p.Address) == 0 {
		return fmt.Errorf("Address cannot be empty")
	}
	if len(p.Validators) == 0 {
		return fmt.Errorf("Validators cannot be empty")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p ConvertSubnetToL1TxParams) GetChainType() string {
	return "P-Chain"
}
