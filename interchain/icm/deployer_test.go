// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package icm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetReleaseURLs(t *testing.T) {
	version := "v1.0.0"
	messengerAddr, deployerAddr, deployerTx, registryBytecode := GetReleaseURLs(version)

	require.Contains(t, messengerAddr, version)
	require.Contains(t, messengerAddr, "TeleporterMessenger_Contract_Address")

	require.Contains(t, deployerAddr, version)
	require.Contains(t, deployerAddr, "TeleporterMessenger_Deployer_Address")

	require.Contains(t, deployerTx, version)
	require.Contains(t, deployerTx, "TeleporterMessenger_Deployment_Transaction")

	require.Contains(t, registryBytecode, version)
	require.Contains(t, registryBytecode, "TeleporterRegistry_Bytecode")
}

func TestDeployer_SetAndGet(t *testing.T) {
	d := &Deployer{}

	messengerAddr := "0x1234567890123456789012345678901234567890"
	deployerAddr := "0x0987654321098765432109876543210987654321"
	deployerTx := []byte("deployment transaction")
	registryBytecode := []byte("registry bytecode")

	d.Set(messengerAddr, deployerAddr, deployerTx, registryBytecode)

	gotMessengerAddr, gotDeployerAddr, gotDeployerTx, gotRegistryBytecode := d.Get()

	require.Equal(t, messengerAddr, gotMessengerAddr)
	require.Equal(t, deployerAddr, gotDeployerAddr)
	require.Equal(t, deployerTx, gotDeployerTx)
	require.Equal(t, registryBytecode, gotRegistryBytecode)
}

func TestDeployer_Validate(t *testing.T) {
	tests := []struct {
		name             string
		messengerAddr    string
		deployerAddr     string
		deployerTx       []byte
		registryBytecode []byte
		wantErr          bool
	}{
		{
			name:             "all fields set",
			messengerAddr:    "0x1234567890123456789012345678901234567890",
			deployerAddr:     "0x0987654321098765432109876543210987654321",
			deployerTx:       []byte("tx"),
			registryBytecode: []byte("bytecode"),
			wantErr:          false,
		},
		{
			name:             "missing messenger address",
			messengerAddr:    "",
			deployerAddr:     "0x0987654321098765432109876543210987654321",
			deployerTx:       []byte("tx"),
			registryBytecode: []byte("bytecode"),
			wantErr:          true,
		},
		{
			name:             "missing deployer address",
			messengerAddr:    "0x1234567890123456789012345678901234567890",
			deployerAddr:     "",
			deployerTx:       []byte("tx"),
			registryBytecode: []byte("bytecode"),
			wantErr:          true,
		},
		{
			name:             "missing deployer tx",
			messengerAddr:    "0x1234567890123456789012345678901234567890",
			deployerAddr:     "0x0987654321098765432109876543210987654321",
			deployerTx:       []byte{},
			registryBytecode: []byte("bytecode"),
			wantErr:          true,
		},
		{
			name:             "missing registry bytecode",
			messengerAddr:    "0x1234567890123456789012345678901234567890",
			deployerAddr:     "0x0987654321098765432109876543210987654321",
			deployerTx:       []byte("tx"),
			registryBytecode: []byte{},
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployer{}
			d.Set(tt.messengerAddr, tt.deployerAddr, tt.deployerTx, tt.registryBytecode)

			err := d.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeployer_LoadDefault(t *testing.T) {
	d := &Deployer{}
	d.LoadDefault()

	require.NoError(t, d.Validate())

	messengerAddr, deployerAddr, deployerTx, registryBytecode := d.Get()

	require.Equal(t, DefaultMessengerContractAddress, messengerAddr)
	require.Equal(t, DefaultMessengerDeployerAddress, deployerAddr)
	require.NotEmpty(t, deployerTx)
	require.NotEmpty(t, registryBytecode)
}

func TestDeployer_LoadFromFiles(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test files
	messengerAddrPath := tmpDir + "/messenger_addr.txt"
	deployerAddrPath := tmpDir + "/deployer_addr.txt"
	deployerTxPath := tmpDir + "/deployer_tx.bin"
	registryBydecodePath := tmpDir + "/registry.bin"

	messengerAddr := "0x1234567890123456789012345678901234567890"
	deployerAddr := "0x0987654321098765432109876543210987654321"
	deployerTx := []byte("deployment transaction bytes")
	registryBytecode := []byte("registry bytecode bytes")

	require.NoError(t, os.WriteFile(messengerAddrPath, []byte(messengerAddr), 0o600))
	require.NoError(t, os.WriteFile(deployerAddrPath, []byte(deployerAddr), 0o600))
	require.NoError(t, os.WriteFile(deployerTxPath, deployerTx, 0o600))
	require.NoError(t, os.WriteFile(registryBydecodePath, registryBytecode, 0o600))

	// Test loading from files
	d := &Deployer{}
	err := d.LoadFromFiles(messengerAddrPath, deployerAddrPath, deployerTxPath, registryBydecodePath)
	require.NoError(t, err)

	gotMessengerAddr, gotDeployerAddr, gotDeployerTx, gotRegistryBytecode := d.Get()

	require.Equal(t, messengerAddr, gotMessengerAddr)
	require.Equal(t, deployerAddr, gotDeployerAddr)
	require.Equal(t, deployerTx, gotDeployerTx)
	require.Equal(t, registryBytecode, gotRegistryBytecode)
}

func TestDeployer_LoadFromFiles_MissingFile(t *testing.T) {
	d := &Deployer{}
	err := d.LoadFromFiles(
		"/nonexistent/messenger.txt",
		"/nonexistent/deployer.txt",
		"/nonexistent/tx.bin",
		"/nonexistent/registry.bin",
	)
	require.Error(t, err)
}
