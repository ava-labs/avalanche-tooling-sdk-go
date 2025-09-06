// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"github.com/ava-labs/avalanchego/utils/constants"
)

type NetworkKind int64

const (
	Undefined NetworkKind = iota
	Mainnet
	Fuji
	Devnet
)

const (
	FujiAPIEndpoint    = "https://api.avax-test.network"
	MainnetAPIEndpoint = "https://api.avax.network"
)

type Network struct {
	Kind     NetworkKind
	ID       uint32
	Endpoint string
}

var UndefinedNetwork = Network{}

func NetworkFromNetworkID(networkID uint32) Network {
	switch networkID {
	case constants.MainnetID:
		return MainnetNetwork()
	case constants.FujiID:
		return FujiNetwork()
	}
	return UndefinedNetwork
}

func NewNetwork(kind NetworkKind, id uint32, endpoint string) Network {
	return Network{
		Kind:     kind,
		ID:       id,
		Endpoint: endpoint,
	}
}

func FujiNetwork() Network {
	return NewNetwork(Fuji, constants.FujiID, FujiAPIEndpoint)
}

func MainnetNetwork() Network {
	return NewNetwork(Mainnet, constants.MainnetID, MainnetAPIEndpoint)
}

// HRPFromNetworkID returns the Human Readable Part (HRP) for a given NetworkID.
// This function maps all known Avalanche network IDs to their corresponding HRP values.
// For unknown NetworkIDs, it returns the FallbackHRP.
func HRPFromNetworkID(networkID uint32) string {
	switch networkID {
	case constants.MainnetID:
		return constants.MainnetHRP
	case constants.CascadeID:
		return constants.CascadeHRP
	case constants.DenaliID:
		return constants.DenaliHRP
	case constants.EverestID:
		return constants.EverestHRP
	case constants.FujiID:
		return constants.FujiHRP
	case constants.UnitTestID:
		return constants.UnitTestHRP
	case constants.LocalID:
		return constants.LocalHRP
	default:
		return constants.FallbackHRP
	}
}
