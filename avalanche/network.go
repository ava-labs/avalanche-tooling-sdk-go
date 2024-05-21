// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

type NetworkKind int64

const (
	Undefined NetworkKind = iota
	Mainnet
	Fuji
	LocalNetwork
	Devnet
)

type Network struct {
	Kind NetworkKind

	ID uint32

	Endpoint string

	ClusterName string
}
