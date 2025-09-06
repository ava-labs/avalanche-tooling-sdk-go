// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/stretchr/testify/require"
)

func TestHRPFromNetworkID(t *testing.T) {
	tests := []struct {
		name        string
		networkID   uint32
		expectedHRP string
	}{
		{
			name:        "Mainnet",
			networkID:   constants.MainnetID,
			expectedHRP: constants.MainnetHRP,
		},
		{
			name:        "Fuji",
			networkID:   constants.FujiID,
			expectedHRP: constants.FujiHRP,
		},
		{
			name:        "Cascade",
			networkID:   constants.CascadeID,
			expectedHRP: constants.CascadeHRP,
		},
		{
			name:        "Denali",
			networkID:   constants.DenaliID,
			expectedHRP: constants.DenaliHRP,
		},
		{
			name:        "Everest",
			networkID:   constants.EverestID,
			expectedHRP: constants.EverestHRP,
		},
		{
			name:        "UnitTest",
			networkID:   constants.UnitTestID,
			expectedHRP: constants.UnitTestHRP,
		},
		{
			name:        "Local",
			networkID:   constants.LocalID,
			expectedHRP: constants.LocalHRP,
		},
		{
			name:        "Unknown NetworkID uses FallbackHRP",
			networkID:   99999,
			expectedHRP: constants.FallbackHRP,
		},
		{
			name:        "Zero NetworkID uses FallbackHRP",
			networkID:   0,
			expectedHRP: constants.FallbackHRP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HRPFromNetworkID(tt.networkID)
			require.Equal(t, tt.expectedHRP, result)
		})
	}
}

func TestHRPFromNetworkID_AllKnownNetworks(t *testing.T) {
	// Test all known network constants to ensure they map correctly
	knownNetworks := map[uint32]string{
		constants.MainnetID:  constants.MainnetHRP,
		constants.CascadeID:  constants.CascadeHRP,
		constants.DenaliID:   constants.DenaliHRP,
		constants.EverestID:  constants.EverestHRP,
		constants.FujiID:     constants.FujiHRP,
		constants.UnitTestID: constants.UnitTestHRP,
		constants.LocalID:    constants.LocalHRP,
	}

	for networkID, expectedHRP := range knownNetworks {
		t.Run(expectedHRP, func(t *testing.T) {
			result := HRPFromNetworkID(networkID)
			require.Equal(t, expectedHRP, result)
		})
	}
}

func TestHRPFromNetworkID_EdgeCases(t *testing.T) {
	// Test edge cases and ensure they return FallbackHRP
	edgeCases := []uint32{
		0,          // Zero
		4294967295, // Max uint32
		123456,     // Random large number
		1000000,    // Another random number
	}

	for _, networkID := range edgeCases {
		t.Run(fmt.Sprintf("NetworkID_%d", networkID), func(t *testing.T) {
			result := HRPFromNetworkID(networkID)
			require.Equal(t, constants.FallbackHRP, result)
		})
	}
}
