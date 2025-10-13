// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"

	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

// ConvertSubnetToL1TxParams contains all parameters needed to create a ConvertSubnetToL1Tx
type ConvertSubnetToL1TxParams struct {
	// SubnetAuthKeys are the keys used to sign `ConvertSubnetToL1Tx`
	SubnetAuthKeys []ids.ShortID
	// SubnetID is Subnet ID of the subnet to convert to an L1.
	SubnetID ids.ID
	// ChainID is Blockchain ID of the L1 where the validator manager contract is deployed.
	ChainID ids.ID
	// Address is address of the validator manager contract.
	Address []byte
	// Validators are the initial set of L1 validators after the conversion.
	Validators []*avagoTxs.ConvertSubnetToL1Validator
}

// Validate validates the parameters
func (p ConvertSubnetToL1TxParams) Validate() error {
	if p.SubnetID == ids.Empty {
		return fmt.Errorf("SubnetID cannot be empty")
	}
	if p.ChainID == ids.Empty {
		return fmt.Errorf("ChainID cannot be empty")
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
	return "P-Chain"
}
