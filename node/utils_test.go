// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultProjectNameFromGCPCredentials(t *testing.T) {
	// Create a temporary file to act as the credentials file
	tempFile, err := os.CreateTemp("", "credentials-*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write a sample JSON configuration to the temporary file
	// #nosec G101
	sampleConfig := `{
		"client_id": "test-client-id",
		"client_secret": "test-client-secret",
		"quota_project_id": "test-project-id",
		"refresh_token": "test-refresh-token",
		"type": "service_account"
	}`
	_, err = tempFile.WriteString(sampleConfig)
	require.NoError(t, err)
	_ = tempFile.Close()

	// Test the function with a valid credentials file
	projectID, err := getDefaultProjectNameFromGCPCredentials(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, "test-project-id", projectID)

	// Test the function with a non-existent file
	_, err = getDefaultProjectNameFromGCPCredentials("nonexistent-file.json")
	require.Error(t, err)

	// Test the function with an invalid JSON file
	invalidFile, err := os.CreateTemp("", "invalid-*.json")
	require.NoError(t, err)
	defer os.Remove(invalidFile.Name())

	_, err = invalidFile.WriteString("invalid json")
	require.NoError(t, err)
	_ = invalidFile.Close()

	_, err = getDefaultProjectNameFromGCPCredentials(invalidFile.Name())
	require.Error(t, err)
}

func TestGetPublicKeyFromSSHKey(t *testing.T) {
	// Create temporary directory to store test SSH key files
	tempDir := t.TempDir()

	// Create a test SSH public key file
	testPublicKeyPath := tempDir + "/id_rsa.pub"
	testPublicKeyContent := "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAr8E7T/ZoQ9Jyb5F1U1t/9F+nkRoSi8g8j6x0g7vZJ68dVVpREzK84+R5cOJ6ydP9Nd+G99kW1HLhfwK5BhJnW3uZ7h1mL0Hh/RZb8csViNe8sEc2FSgH5G8cl3ZX8Y1UtdbS4k5F3cC3B4JFF9y6vOZRwUBO4z1Z2BZaGP29sXXkW0ZGRrWaBswcq+S5FJ1QOeeJ38OjkB45L7zq2X2NQ== user@hostname"
	err := os.WriteFile(testPublicKeyPath, []byte(testPublicKeyContent+"\n"), 0o600)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name     string
		keyPath  string
		expected string
		wantErr  bool
	}{
		{
			name:     "ValidCustomPath",
			keyPath:  testPublicKeyPath,
			expected: testPublicKeyContent,
			wantErr:  false,
		},
		{
			name:     "InvalidPath",
			keyPath:  tempDir + "/nonexistent.pub",
			expected: "",
			wantErr:  true,
		},
	}

	// Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPublicKeyFromSSHKey(tt.keyPath)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, got)
			}
		})
	}
}
