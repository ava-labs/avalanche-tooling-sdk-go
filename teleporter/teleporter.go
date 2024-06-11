// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package teleporter

import "math/big"

type Info struct {
	Version                  string
	FundedAddress            string
	FundedBalance            *big.Int
	MessengerDeployerAddress string
	RelayerAddress           string
}
