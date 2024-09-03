// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	remoteconfig "github.com/ava-labs/avalanche-tooling-sdk-go/node/config"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// ValidateComposeFile validates a docker-compose file on a remote node.
func (h *Node) ValidateComposeFile(composeFile string, timeout time.Duration) error {
	if output, err := h.Commandf(nil, timeout, "docker compose -f %s config", composeFile); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// ComposeSSHSetupNode sets up an AvalancheGo node and dependencies on a remote node over SSH.
func (h *Node) ComposeSSHSetupNode(networkID string, subnetsToTrack []string, avalancheGoVersion string, withMonitoring bool) error {
	startTime := time.Now()
	folderStructure := remoteconfig.RemoteFoldersToCreateAvalanchego()
	for _, dir := range folderStructure {
		if err := h.MkdirAll(dir, constants.SSHFileOpsTimeout); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	h.Logger.Infof("avalancheCLI folder structure created on remote node %s after %s", folderStructure, time.Since(startTime))
	avagoDockerImage := fmt.Sprintf("%s:%s", constants.AvalancheGoDockerImage, avalancheGoVersion)
	h.Logger.Infof("Preparing AvalancheGo Docker image %s on %s[%s]", avagoDockerImage, h.NodeID, h.IP)
	if err := h.PrepareDockerImageWithRepo(avagoDockerImage, constants.AvalancheGoGitRepo, avalancheGoVersion); err != nil {
		return err
	}
	h.Logger.Infof("AvalancheGo Docker image %s ready on %s[%s] after %s", avagoDockerImage, h.NodeID, h.IP, time.Since(startTime))
	if err := h.RunSSHRenderAvalancheNodeConfig(networkID, subnetsToTrack); err != nil {
		return err
	}
	h.Logger.Infof("AvalancheGo configs uploaded to %s[%s] after %s", h.NodeID, h.IP, time.Since(startTime))
	return h.ComposeOverSSH("Compose Node",
		constants.SSHScriptTimeout,
		"templates/avalanchego.docker-compose.yml",
		dockerComposeInputs{
			AvalanchegoVersion: avalancheGoVersion,
			WithMonitoring:     withMonitoring,
			WithAvalanchego:    true,
			E2E:                utils.IsE2E(),
			E2EIP:              utils.E2EConvertIP(h.IP),
			E2ESuffix:          utils.E2ESuffix(h.IP),
		})
}

func (h *Node) ComposeSSHSetupLoadTest() error {
	return h.ComposeOverSSH("Compose Node",
		constants.SSHScriptTimeout,
		"templates/avalanchego.docker-compose.yml",
		dockerComposeInputs{
			WithMonitoring:  true,
			WithAvalanchego: false,
		})
}

// WasNodeSetupWithMonitoring checks if an AvalancheGo node was setup with monitoring on a remote node.
func (h *Node) WasNodeSetupWithMonitoring() (bool, error) {
	return h.HasRemoteComposeService(utils.GetRemoteComposeFile(), constants.ServicePromtail, constants.SSHScriptTimeout)
}

// ComposeSSHSetupMonitoring sets up monitoring using docker-compose.
func (h *Node) ComposeSSHSetupMonitoring() error {
	grafanaConfigFile, grafanaDashboardsFile, grafanaLokiDatasourceFile, grafanaPromDatasourceFile, err := prepareGrafanaConfig()
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(grafanaLokiDatasourceFile); err != nil {
			h.Logger.Errorf("Error removing temporary file %s: %s", grafanaLokiDatasourceFile, err)
		}
		if err := os.Remove(grafanaPromDatasourceFile); err != nil {
			h.Logger.Errorf("Error removing temporary file %s: %s", grafanaPromDatasourceFile, err)
		}
		if err := os.Remove(grafanaDashboardsFile); err != nil {
			h.Logger.Errorf("Error removing temporary file %s: %s", grafanaDashboardsFile, err)
		}
		if err := os.Remove(grafanaConfigFile); err != nil {
			h.Logger.Errorf("Error removing temporary file %s: %s", grafanaConfigFile, err)
		}
	}()

	grafanaLokiDatasourceRemoteFileName := filepath.Join(utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "provisioning", "datasources"), "loki.yml")
	if err := h.Upload(grafanaLokiDatasourceFile, grafanaLokiDatasourceRemoteFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	grafanaPromDatasourceFileName := filepath.Join(utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "provisioning", "datasources"), "prometheus.yml")
	if err := h.Upload(grafanaPromDatasourceFile, grafanaPromDatasourceFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	grafanaDashboardsRemoteFileName := filepath.Join(utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "provisioning", "dashboards"), "dashboards.yml")
	if err := h.Upload(grafanaDashboardsFile, grafanaDashboardsRemoteFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	grafanaConfigRemoteFileName := filepath.Join(utils.GetRemoteComposeServicePath(constants.ServiceGrafana), "grafana.ini")
	if err := h.Upload(grafanaConfigFile, grafanaConfigRemoteFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}

	return h.ComposeOverSSH("Setup Monitoring",
		constants.SSHScriptTimeout,
		"templates/monitoring.docker-compose.yml",
		dockerComposeInputs{})
}

func (h *Node) ComposeSSHSetupAWMRelayer() error {
	return h.ComposeOverSSH("Setup AWM Relayer",
		constants.SSHScriptTimeout,
		"templates/awmrelayer.docker-compose.yml",
		dockerComposeInputs{})
}
