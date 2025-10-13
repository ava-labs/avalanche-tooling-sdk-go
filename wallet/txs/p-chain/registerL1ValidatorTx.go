// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// RegisterL1ValidatorParams contains all parameters needed to create a ConvertSubnetToL1Tx
type RegisterL1ValidatorParams struct {
	Balance              uint64
	BLSPublicKey         string
	BLSProofOfPossession string
	Message              string
}

// GetTxType returns the transaction type identifier
func (p RegisterL1ValidatorParams) GetTxType() string {
	return constants.PChainRegisterL1ValidatorTx
}

// Validate validates the parameters
func (p RegisterL1ValidatorParams) Validate() error {
	if p.Balance == 0 {
		return fmt.Errorf("subnet auth keys cannot be empty")
	}
	if p.BLSPublicKey == "" {
		return fmt.Errorf("subnet ID cannot be empty")
	}
	if p.BLSProofOfPossession == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if p.BLSProofOfPossession == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p RegisterL1ValidatorParams) GetChainType() string {
	return constants.ChainTypePChain
}
