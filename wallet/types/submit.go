// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import "fmt"

// SubmitTxParams contains parameters for submitting transactions (build, sign, and send)
type SubmitTxParams struct {
	AccountNames []string
	BuildTxInput
}

// Validate validates the submit transaction parameters
func (p *SubmitTxParams) Validate() error {
	if len(p.AccountNames) > 1 {
		return fmt.Errorf("only one account name is currently supported")
	}
	if p.BuildTxInput == nil {
		return fmt.Errorf("build tx input is required")
	}
	return p.BuildTxInput.Validate()
}

// SubmitTxResult represents the result of submitting a transaction
type SubmitTxResult struct {
	SendTxOutput
}

// Validate validates the submit transaction result
func (r *SubmitTxResult) Validate() error {
	if r.SendTxOutput == nil {
		return fmt.Errorf("send tx output is required")
	}
	return r.SendTxOutput.Validate()
}
