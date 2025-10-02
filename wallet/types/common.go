// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package types

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/subnet-evm/ethclient"
)

type ChainClients struct {
	C *ethclient.Client // …/ext/bc/C/rpc
	X string            // …/ext/bc/X
	P string            // …/ext/bc/P
}

// BaseParams contains common parameters shared across all operations
type BaseParams struct {
	Account account.Account
	Network network.Network
}

// Validate validates the base parameters
func (p *BaseParams) Validate() error {
	if p.Account == nil {
		return fmt.Errorf("account is required")
	}
	if p.Network.Kind == network.Undefined {
		return fmt.Errorf("network is required")
	}
	return nil
}
