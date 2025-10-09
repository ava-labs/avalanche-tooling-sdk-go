// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// CreateSubnetTxParams contains all parameters needed to create a ConvertSubnetToL1Tx
type CreateSubnetTxParams struct {
	ControlKeys []string
	Threshold   uint32
}

// GetTxType returns the transaction type identifier
func (p CreateSubnetTxParams) GetTxType() string {
	return constants.PChainCreateSubnetTx
}

// Validate validates the parameters
func (p CreateSubnetTxParams) Validate() error {
	if p.ControlKeys == nil {
		return fmt.Errorf("control keys cannot be empty")
	}
	if p.Threshold == 0 {
		return fmt.Errorf("threshold cannot be zero")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p CreateSubnetTxParams) GetChainType() string {
	return constants.ChainTypePChain
}
