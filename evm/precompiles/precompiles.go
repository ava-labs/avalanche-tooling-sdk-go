// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package precompiles

import (
	"github.com/ava-labs/subnet-evm/precompile/contracts/nativeminter"
	"github.com/ava-labs/subnet-evm/precompile/contracts/warp"
)

var (
	NativeMinterPrecompile = nativeminter.ContractAddress
	WarpPrecompile         = warp.ContractAddress
)
