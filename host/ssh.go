// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	remoteconfig "github.com/ava-labs/avalanche-tooling-sdk-go/host/config"
	"github.com/ava-labs/avalanche-tooling-sdk-go/host/monitoring"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

type scriptInputs struct {
	AvalancheGoVersion   string
	CLIVersion           string
	SubnetExportFileName string
	SubnetName           string
	ClusterName          string
	GoVersion            string
	CliBranch            string
	IsDevNet             bool
	NetworkFlag          string
	SubnetEVMBinaryPath  string
	SubnetEVMReleaseURL  string
	SubnetEVMArchive     string
	LoadTestRepoDir      string
	LoadTestRepo         string
	LoadTestPath         string
	LoadTestCommand      string
	LoadTestBranch       string
	LoadTestGitCommit    string
	CheckoutCommit       bool
	LoadTestResultFile   string
	GrafanaPkg           string
}

//go:embed shell/*.sh
var script embed.FS

// RunOverSSH runs provided script path over ssh.
// This script can be template as it will be rendered using scriptInputs vars
func (h *Host) RunOverSSH(
	scriptDesc string,
	timeout time.Duration,
	scriptPath string,
	templateVars scriptInputs,
) error {
	startTime := time.Now()
	shellScript, err := script.ReadFile(scriptPath)
	if err != nil {
		return err
	}
	var script bytes.Buffer
	t, err := template.New(scriptDesc).Parse(string(shellScript))
	if err != nil {
		return err
	}
	err = t.Execute(&script, templateVars)
	if err != nil {
		return err
	}

	if output, err := h.Command(nil, timeout, script.String()); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	executionTime := time.Since(startTime)
	h.Logger.Infof("RunOverSSH[%s]%s took %s with err: %v", h.NodeID, scriptDesc, executionTime, err)
	return nil
}

// RunSSHSetupNode runs script to setup sdk dependencies on a remote host over SSH.
func (h *Host) RunSSHSetupNode(cliVersion string) error {
	if err := h.RunOverSSH(
		"Setup Node",
		constants.SSHLongRunningScriptTimeout,
		"shell/setupNode.sh",
		scriptInputs{CLIVersion: cliVersion},
	); err != nil {
		return err
	}
	return nil
}

// RunSSHSetupDockerService runs script to setup docker compose service for CLI
func (h *Host) RunSSHSetupDockerService() error {
	if h.IsSystemD() {
		return h.RunOverSSH(
			"Setup Docker Service",
			constants.SSHLongRunningScriptTimeout,
			"shell/setupDockerService.sh",
			scriptInputs{},
		)
	} else {
		// no need to setup docker service
		return nil
	}
}

// RunSSHRestartNode runs script to restart avalanchego
func (h *Host) RunSSHRestartNode() error {
	remoteComposeFile := utils.GetRemoteComposeFile()
	return h.RestartDockerComposeService(remoteComposeFile, "avalanchego", constants.SSHLongRunningScriptTimeout)
}

// RunSSHStartAWMRelayerService runs script to start an AWM Relayer Service
func (h *Host) RunSSHStartAWMRelayerService() error {
	return h.StartDockerComposeService(utils.GetRemoteComposeFile(), "awm-relayer", constants.SSHLongRunningScriptTimeout)
}

// RunSSHStopAWMRelayerService runs script to start an AWM Relayer Service
func (h *Host) RunSSHStopAWMRelayerService() error {
	return h.StopDockerComposeService(utils.GetRemoteComposeFile(), "awm-relayer", constants.SSHLongRunningScriptTimeout)
}

// RunSSHUpgradeAvalanchego runs script to upgrade avalanchego
func (h *Host) RunSSHUpgradeAvalanchego(networkID string, avalancheGoVersion string) error {
	withMonitoring, err := h.WasNodeSetupWithMonitoring()
	if err != nil {
		return err
	}

	if err := h.ComposeSSHSetupNode(networkID, avalancheGoVersion, withMonitoring); err != nil {
		return err
	}
	return h.RestartDockerCompose(constants.SSHLongRunningScriptTimeout)
}

// RunSSHStartNode runs script to start avalanchego
func (h *Host) RunSSHStartNode() error {
	return h.StartDockerComposeService(utils.GetRemoteComposeFile(), "avalanchego", constants.SSHLongRunningScriptTimeout)
}

// RunSSHStopNode runs script to stop avalanchego
func (h *Host) RunSSHStopNode() error {
	return h.StopDockerComposeService(utils.GetRemoteComposeFile(), "avalanchego", constants.SSHLongRunningScriptTimeout)
}

// RunSSHUpgradeSubnetEVM runs script to upgrade subnet evm
func (h *Host) RunSSHUpgradeSubnetEVM(subnetEVMBinaryPath string) error {
	return h.RunOverSSH(
		"Upgrade Subnet EVM",
		constants.SSHScriptTimeout,
		"shell/upgradeSubnetEVM.sh",
		scriptInputs{SubnetEVMBinaryPath: subnetEVMBinaryPath},
	)
}

func (h *Host) RunSSHSetupPrometheusConfig(avalancheGoPorts, machinePorts, loadTestPorts []string) error {
	for _, folder := range remoteconfig.PrometheusFoldersToCreate() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	cloudNodePrometheusConfigTemp := utils.GetRemoteComposeServicePath("prometheus", "prometheus.yml")
	promConfig, err := os.CreateTemp("", "prometheus")
	if err != nil {
		return err
	}
	defer os.Remove(promConfig.Name())
	if err := monitoring.WritePrometheusConfig(promConfig.Name(), avalancheGoPorts, machinePorts, loadTestPorts); err != nil {
		return err
	}

	return h.Upload(
		promConfig.Name(),
		cloudNodePrometheusConfigTemp,
		constants.SSHFileOpsTimeout,
	)
}

func (h *Host) RunSSHSetupLokiConfig(port int) error {
	for _, folder := range remoteconfig.LokiFoldersToCreate() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	cloudNodeLokiConfigTemp := utils.GetRemoteComposeServicePath("loki", "loki.yml")
	lokiConfig, err := os.CreateTemp("", "loki")
	if err != nil {
		return err
	}
	defer os.Remove(lokiConfig.Name())
	if err := monitoring.WriteLokiConfig(lokiConfig.Name(), strconv.Itoa(port)); err != nil {
		return err
	}
	return h.Upload(
		lokiConfig.Name(),
		cloudNodeLokiConfigTemp,
		constants.SSHFileOpsTimeout,
	)
}

func (h *Host) RunSSHSetupPromtailConfig(lokiIP string, lokiPort int, cloudID string, nodeID string, chainID string) error {
	for _, folder := range remoteconfig.PromtailFoldersToCreate() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	cloudNodePromtailConfigTemp := utils.GetRemoteComposeServicePath("promtail", "promtail.yml")
	promtailConfig, err := os.CreateTemp("", "promtail")
	if err != nil {
		return err
	}
	defer os.Remove(promtailConfig.Name())

	if err := monitoring.WritePromtailConfig(promtailConfig.Name(), lokiIP, strconv.Itoa(lokiPort), cloudID, nodeID, chainID); err != nil {
		return err
	}
	return h.Upload(
		promtailConfig.Name(),
		cloudNodePromtailConfigTemp,
		constants.SSHFileOpsTimeout,
	)
}

func (h *Host) RunSSHUploadNodeAWMRelayerConfig(nodeInstanceDirPath string) error {
	cloudAWMRelayerConfigDir := filepath.Join(constants.CloudNodeCLIConfigBasePath, constants.ServicesDir, constants.AWMRelayerInstallDir)
	if err := h.MkdirAll(cloudAWMRelayerConfigDir, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	return h.Upload(
		filepath.Join(nodeInstanceDirPath, constants.ServicesDir, constants.AWMRelayerInstallDir, constants.AWMRelayerConfigFilename),
		filepath.Join(cloudAWMRelayerConfigDir, constants.AWMRelayerConfigFilename),
		constants.SSHFileOpsTimeout,
	)
}

// RunSSHGetNewSubnetEVMRelease runs script to download new subnet evm
func (h *Host) RunSSHGetNewSubnetEVMRelease(subnetEVMReleaseURL, subnetEVMArchive string) error {
	return h.RunOverSSH(
		"Get Subnet EVM Release",
		constants.SSHScriptTimeout,
		"shell/getNewSubnetEVMRelease.sh",
		scriptInputs{SubnetEVMReleaseURL: subnetEVMReleaseURL, SubnetEVMArchive: subnetEVMArchive},
	)
}

// RunSSHUploadStakingFiles uploads staking files to a remote host via SSH.
func (h *Host) RunSSHUploadStakingFiles(keyPath string) error {
	if err := h.MkdirAll(
		constants.CloudNodeStakingPath,
		constants.SSHFileOpsTimeout,
	); err != nil {
		return err
	}
	if err := h.Upload(
		filepath.Join(keyPath, constants.StakerCertFileName),
		filepath.Join(constants.CloudNodeStakingPath, constants.StakerCertFileName),
		constants.SSHFileOpsTimeout,
	); err != nil {
		return err
	}
	if err := h.Upload(
		filepath.Join(keyPath, constants.StakerKeyFileName),
		filepath.Join(constants.CloudNodeStakingPath, constants.StakerKeyFileName),
		constants.SSHFileOpsTimeout,
	); err != nil {
		return err
	}
	return h.Upload(
		filepath.Join(keyPath, constants.BLSKeyFileName),
		filepath.Join(constants.CloudNodeStakingPath, constants.BLSKeyFileName),
		constants.SSHFileOpsTimeout,
	)
}
