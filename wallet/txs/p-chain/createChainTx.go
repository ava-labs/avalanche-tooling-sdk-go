// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/types"
)

// CreateChainTxParams contains all parameters needed to create a ConvertSubnetToL1Tx
type CreateChainTxParams struct {
	// SubnetAuthKeys are the keys used to sign `ConvertSubnetToL1Tx`
	SubnetAuthKeys []string
	// SubnetID specifies the subnet to launch the chain in
	SubnetID string
	// VMID specifies the vm that the new chain will run.
	VMID string
	// ChainName specifies a human-readable name for the chain
	ChainName string
	// Genesis specifies the initial state of the new chain
	Genesis []byte
}

// GetTxType returns the transaction type identifier
func (p CreateChainTxParams) GetTxType() string {
	return types.PChainCreateChainTx
}

// Validate validates the parameters
func (p CreateChainTxParams) Validate() error {
	if p.SubnetAuthKeys == nil {
		return fmt.Errorf("subnet auth keys cannot be empty")
	}
	if p.SubnetID == "" {
		return fmt.Errorf("subnet ID cannot be empty")
	}
	if p.VMID == "" {
		return fmt.Errorf("VMID cannot be empty")
	}
	if p.ChainName == "" {
		return fmt.Errorf("chain name cannot be empty")
	}
	if len(p.Genesis) == 0 {
		return fmt.Errorf("genesis cannot be empty")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p CreateChainTxParams) GetChainType() string {
	return types.ChainTypePChain
}
