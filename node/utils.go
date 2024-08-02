// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	remoteconfig "github.com/ava-labs/avalanche-tooling-sdk-go/node/config"
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

// isMonitoringNode checks if the node has the Monitor role.
//
// Parameter(s):
// - node *Node: The node to check.
// Return type(s): bool
func isMonitoringNode(node Node) bool {
	return slices.Contains(node.Roles, Monitor)
}

// isAvalancheGoNode checks if the node has the API or Validator role.
//
// - node *Node: The node to check.
// bool
func isAvalancheGoNode(node Node) bool {
	return slices.Contains(node.Roles, API) || slices.Contains(node.Roles, Validator)
}

// isLoadTestNode checks if the node has the LoadTest role.
//
// - node *Node: The node to check.
// bool
func isLoadTestNode(node Node) bool {
	return slices.Contains(node.Roles, Loadtest)
}

// getPrometheusTargets returns the Prometheus targets for the given nodes.
//
// Parameters:
// - nodes: a slice of Node representing the nodes to get the Prometheus targets for.
//
// Returns:
// - avalancheGoPorts: a slice of strings representing the Prometheus targets for the nodes with the AvalancheGo role.
// - machinePorts: a slice of strings representing the Prometheus targets for the nodes with the AvalancheGo role.
// - ltPorts: a slice of strings representing the Prometheus targets for the nodes with the LoadTest role.
func getPrometheusTargets(nodes []Node) ([]string, []string, []string) {
	avalancheGoPorts := []string{}
	machinePorts := []string{}
	ltPorts := []string{}
	for _, host := range nodes {
		if isAvalancheGoNode(host) {
			avalancheGoPorts = append(avalancheGoPorts, fmt.Sprintf("'%s:%s'", host.IP, strconv.Itoa(constants.AvalanchegoAPIPort)))
			machinePorts = append(machinePorts, fmt.Sprintf("'%s:%s'", host.IP, strconv.Itoa(constants.AvalanchegoMachineMetricsPort)))
		}
		if isLoadTestNode(host) {
			ltPorts = append(ltPorts, fmt.Sprintf("'%s:%s'", host.IP, strconv.Itoa(constants.AvalanchegoLoadTestPort)))
		}
	}
	return avalancheGoPorts, machinePorts, ltPorts
}

func composeFileExists(node Node) bool {
	composeFileExists, _ := node.FileExists(utils.GetRemoteComposeFile())
	return composeFileExists
}

func genesisFileExists(node Node) bool {
	genesisFileExists, _ := node.FileExists(remoteconfig.GetRemoteAvalancheGenesis())
	return genesisFileExists
}

func nodeConfigFileExists(node Node) bool {
	nodeConfigFileExists, _ := node.FileExists(remoteconfig.GetRemoteAvalancheNodeConfig())
	return nodeConfigFileExists
}
