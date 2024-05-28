// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import "github.com/ava-labs/avalanchego/utils/constants"

type NetworkKind int64

type Network struct {
	Kind     NetworkKind
	ID       uint32
	Endpoint string
}

func (n Network) HRP() string {
	switch n.ID {
	case constants.LocalID:
		return constants.LocalHRP
	case constants.FujiID:
		return constants.FujiHRP
	case constants.MainnetID:
		return constants.MainnetHRP
	default:
		return constants.FallbackHRP
	}
}
