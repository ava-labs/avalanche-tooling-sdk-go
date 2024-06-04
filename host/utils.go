// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

const (
	maxResponseSize      = 102400          // 100KB should be enough to read the avalanchego response
	sshConnectionTimeout = 3 * time.Second // usually takes less than 2
	sshConnectionRetries = 5
)

// getDefaultProjectNameFromGCPCredentials returns the default GCP project name
func getDefaultProjectNameFromGCPCredentials(credentialsFilePath string) (string, error) {
	type GCPConfig struct {
		ClientID       string `json:"client_id"`
		ClientSecret   string `json:"client_secret"`
		QuotaProjectID string `json:"quota_project_id"`
		RefreshToken   string `json:"refresh_token"`
		Type           string `json:"type"`
	}
	file, err := os.Open(utils.ExpandHome(credentialsFilePath))
	if err != nil {
		return "", err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	var gcpConfig GCPConfig
	if err := json.Unmarshal(bytes, &gcpConfig); err != nil {
		return "", err
	}
	return gcpConfig.QuotaProjectID, nil
}

// GetPublicKeyFromDefaultSSHKey returns the public key from the default SSH key
func GetPublicKeyFromSSHKey(keyPath string) (string, error) {
	if keyPath == "" {
		keyPath = utils.ExpandHome("~/.ssh/id_rsa.pub")
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(key), "\n"), nil
}
