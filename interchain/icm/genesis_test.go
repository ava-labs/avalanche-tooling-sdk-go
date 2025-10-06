// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package icm

import (
	"math/big"
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/stretchr/testify/require"
)

func TestAddMessengerContractToAllocations(t *testing.T) {
	allocs := make(core.GenesisAlloc)

	AddMessengerContractToAllocations(allocs)

	// Check messenger contract allocation
	messengerAddr := common.HexToAddress(DefaultMessengerContractAddress)
	messengerAlloc, ok := allocs[messengerAddr]
	require.True(t, ok, "messenger contract should be allocated")
	require.NotNil(t, messengerAlloc.Code, "messenger contract should have code")
	require.NotEmpty(t, messengerAlloc.Code, "messenger contract code should not be empty")
	require.Equal(t, uint64(1), messengerAlloc.Nonce)
	require.Equal(t, big.NewInt(0), messengerAlloc.Balance)
	require.NotEmpty(t, messengerAlloc.Storage, "messenger contract should have storage")

	// Check deployer allocation
	deployerAddr := common.HexToAddress(DefaultMessengerDeployerAddress)
	deployerAlloc, ok := allocs[deployerAddr]
	require.True(t, ok, "deployer address should be allocated")
	require.Equal(t, uint64(1), deployerAlloc.Nonce)
	require.Equal(t, big.NewInt(0), deployerAlloc.Balance)
}

func TestAddRegistryContractToAllocations(t *testing.T) {
	allocs := make(core.GenesisAlloc)

	err := AddRegistryContractToAllocations(allocs)
	require.NoError(t, err)

	// Check registry contract allocation
	registryAddr := common.HexToAddress(RegistryContractAddressAtGenesis)
	registryAlloc, ok := allocs[registryAddr]
	require.True(t, ok, "registry contract should be allocated")
	require.NotNil(t, registryAlloc.Code, "registry contract should have code")
	require.NotEmpty(t, registryAlloc.Code, "registry contract code should not be empty")
	require.Equal(t, uint64(1), registryAlloc.Nonce)
	require.Equal(t, big.NewInt(0), registryAlloc.Balance)
	require.NotEmpty(t, registryAlloc.Storage, "registry contract should have storage")
}

func TestAddBothContractsToAllocations(t *testing.T) {
	allocs := make(core.GenesisAlloc)

	AddMessengerContractToAllocations(allocs)
	err := AddRegistryContractToAllocations(allocs)
	require.NoError(t, err)

	// Verify both contracts are allocated
	require.Len(t, allocs, 3) // messenger, deployer, registry

	messengerAddr := common.HexToAddress(DefaultMessengerContractAddress)
	_, ok := allocs[messengerAddr]
	require.True(t, ok, "messenger contract should be allocated")

	deployerAddr := common.HexToAddress(DefaultMessengerDeployerAddress)
	_, ok = allocs[deployerAddr]
	require.True(t, ok, "deployer address should be allocated")

	registryAddr := common.HexToAddress(RegistryContractAddressAtGenesis)
	_, ok = allocs[registryAddr]
	require.True(t, ok, "registry contract should be allocated")
}

func TestHexFill32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short hex",
			input:    "0x1",
			expected: "0000000000000000000000000000000000000000000000000000000000000001",
		},
		{
			name:     "already 32 bytes",
			input:    "0x0000000000000000000000000000000000000000000000000000000000000001",
			expected: "0000000000000000000000000000000000000000000000000000000000000001",
		},
		{
			name:     "no 0x prefix",
			input:    "1",
			expected: "0000000000000000000000000000000000000000000000000000000000000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexFill32(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
