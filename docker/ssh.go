// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ava-labs/avalanche-cli/pkg/constants"
	"github.com/ava-labs/avalanche-cli/pkg/models"
	"github.com/ava-labs/avalanche-cli/pkg/remoteconfig"
	"github.com/ava-labs/avalanche-cli/pkg/utils"
	"github.com/ava-labs/avalanche-cli/pkg/ux"
)

// ValidateComposeFile validates a docker-compose file on a remote host.
func (dh *DockerHost) ValidateComposeFile(composeFile string, timeout time.Duration) error {
	if output, err := dh.Host.Commandf(nil, timeout, "docker compose -f %s config", composeFile); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// ComposeSSHSetupNode sets up an AvalancheGo node and dependencies on a remote host over SSH.
func (dh *DockerHost) ComposeSSHSetupNode(network models.Network, avalancheGoVersion string, withMonitoring bool) error {
	startTime := time.Now()
	folderStructure := remoteconfig.RemoteFoldersToCreateAvalanchego()
	for _, dir := range folderStructure {
		if err := dh.Host.MkdirAll(dir, constants.SSHFileOpsTimeout); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	dh.Logger.Infof("avalancheCLI folder structure created on remote host %s after %s", folderStructure, time.Since(startTime))
	// configs
	networkID := network.NetworkIDFlagValue()
	if network.Kind == models.Local || network.Kind == models.Devnet {
		networkID = fmt.Sprintf("%d", network.ID)
	}

	avagoDockerImage := fmt.Sprintf("%s:%s", constants.AvalancheGoDockerImage, avalancheGoVersion)
	dh.Logger.Infof("Preparing AvalancheGo Docker image %s on %s[%s]", avagoDockerImage, dh.Host.NodeID, dh.Host.IP)
	if err := dh.PrepareDockerImageWithRepo(avagoDockerImage, constants.AvalancheGoGitRepo, avalancheGoVersion); err != nil {
		return err
	}
	dh.Logger.Infof("AvalancheGo Docker image %s ready on %s[%s] after %s", avagoDockerImage, dh.Host.NodeID, dh.Host.IP, time.Since(startTime))
	nodeConfFile, cChainConfFile, err := dh.prepareAvalanchegoConfig(networkID)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(nodeConfFile); err != nil {
			ux.Logger.Error("Error removing temporary file %s: %s", nodeConfFile, err)
		}
		if err := os.Remove(cChainConfFile); err != nil {
			ux.Logger.Error("Error removing temporary file %s: %s", cChainConfFile, err)
		}
	}()

	if err := dh.Host.Upload(nodeConfFile, remoteconfig.GetRemoteAvalancheNodeConfig(), constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	if err := dh.Host.Upload(cChainConfFile, remoteconfig.GetRemoteAvalancheCChainConfig(), constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	dh.Logger.Infof("AvalancheGo configs uploaded to %s[%s] after %s", dh.Host.NodeID, dh.Host.IP, time.Since(startTime))
	return dh.ComposeOverSSH("Compose Node",
		constants.SSHScriptTimeout,
		"templates/avalanchego.docker-compose.yml",
		dockerComposeInputs{
			AvalanchegoVersion: avalancheGoVersion,
			WithMonitoring:     withMonitoring,
			WithAvalanchego:    true,
			E2E:                utils.IsE2E(),
			E2EIP:              utils.E2EConvertIP(dh.Host.IP),
			E2ESuffix:          utils.E2ESuffix(dh.Host.IP),
		})
}

func (dh *DockerHost) ComposeSSHSetupLoadTest() error {
	return dh.ComposeOverSSH("Compose Node",
		constants.SSHScriptTimeout,
		"templates/avalanchego.docker-compose.yml",
		dockerComposeInputs{
			WithMonitoring:  true,
			WithAvalanchego: false,
		})
}

// WasNodeSetupWithMonitoring checks if an AvalancheGo node was setup with monitoring on a remote host.
func (dh *DockerHost) WasNodeSetupWithMonitoring() (bool, error) {
	return dh.HasRemoteComposeService(utils.GetRemoteComposeFile(), "promtail", constants.SSHScriptTimeout)
}

// ComposeSSHSetupMonitoring sets up monitoring using docker-compose.
func (dh *DockerHost) ComposeSSHSetupMonitoring() error {
	grafanaConfigFile, grafanaDashboardsFile, grafanaLokiDatasourceFile, grafanaPromDatasourceFile, err := prepareGrafanaConfig()
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(grafanaLokiDatasourceFile); err != nil {
			ux.Logger.Error("Error removing temporary file %s: %s", grafanaLokiDatasourceFile, err)
		}
		if err := os.Remove(grafanaPromDatasourceFile); err != nil {
			ux.Logger.Error("Error removing temporary file %s: %s", grafanaPromDatasourceFile, err)
		}
		if err := os.Remove(grafanaDashboardsFile); err != nil {
			ux.Logger.Error("Error removing temporary file %s: %s", grafanaDashboardsFile, err)
		}
		if err := os.Remove(grafanaConfigFile); err != nil {
			ux.Logger.Error("Error removing temporary file %s: %s", grafanaConfigFile, err)
		}
	}()

	grafanaLokiDatasourceRemoteFileName := filepath.Join(utils.GetRemoteComposeServicePath("grafana", "provisioning", "datasources"), "loki.yml")
	if err := dh.Host.Upload(grafanaLokiDatasourceFile, grafanaLokiDatasourceRemoteFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	grafanaPromDatasourceFileName := filepath.Join(utils.GetRemoteComposeServicePath("grafana", "provisioning", "datasources"), "prometheus.yml")
	if err := dh.Host.Upload(grafanaPromDatasourceFile, grafanaPromDatasourceFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	grafanaDashboardsRemoteFileName := filepath.Join(utils.GetRemoteComposeServicePath("grafana", "provisioning", "dashboards"), "dashboards.yml")
	if err := dh.Host.Upload(grafanaDashboardsFile, grafanaDashboardsRemoteFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	grafanaConfigRemoteFileName := filepath.Join(utils.GetRemoteComposeServicePath("grafana"), "grafana.ini")
	if err := dh.Host.Upload(grafanaConfigFile, grafanaConfigRemoteFileName, constants.SSHFileOpsTimeout); err != nil {
		return err
	}

	return dh.ComposeOverSSH("Setup Monitoring",
		constants.SSHScriptTimeout,
		"templates/monitoring.docker-compose.yml",
		dockerComposeInputs{})
}

func (dh *DockerHost) ComposeSSHSetupAWMRelayer() error {
	return dh.ComposeOverSSH("Setup AWM Relayer",
		constants.SSHScriptTimeout,
		"templates/awmrelayer.docker-compose.yml",
		dockerComposeInputs{})
}
