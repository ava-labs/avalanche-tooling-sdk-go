// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// RegisterL1ValidatorTxParams contains all parameters needed to create a RegisterL1ValidatorTx
type RegisterL1ValidatorTxParams struct {
	// Balance <= sum($AVAX inputs) - sum($AVAX outputs) - TxFee.
	Balance uint64
	// [Signer] is the BLS key for this validator.
	// Note: We do not enforce that the BLS key is unique across all validators.
	//       This means that validators can share a key if they so choose.
	//       However, a NodeID + Subnet does uniquely map to a BLS key
	BLSPublicKey         string
	BLSProofOfPossession string
	// Message is expected to be a signed Warp message containing an
	// AddressedCall payload with the RegisterL1Validator message.
	Message string
}

// Validate validates the parameters
func (p RegisterL1ValidatorTxParams) Validate() error {
	if p.Balance == 0 {
		return fmt.Errorf("balance cannot be empty")
	}
	if p.BLSPublicKey == "" {
		return fmt.Errorf("bls public key cannot be empty")
	}
	if p.BLSProofOfPossession == "" {
		return fmt.Errorf("bls proof of possession cannot be empty")
	}
	if p.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	return nil
}

// GetChainType returns which chain this transaction is for
func (p RegisterL1ValidatorTxParams) GetChainType() string {
	return constants.ChainTypePChain
}
