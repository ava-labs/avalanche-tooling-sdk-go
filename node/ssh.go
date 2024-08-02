// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	remoteconfig "github.com/ava-labs/avalanche-tooling-sdk-go/node/config"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node/monitoring"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

type scriptInputs struct {
	AvalancheGoVersion   string
	SubnetExportFileName string
	SubnetName           string
	ClusterName          string
	GoVersion            string
	IsDevNet             bool
	NetworkFlag          string
	SubnetVMBinaryPath   string
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
func (h *Node) RunOverSSH(
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
func (h *Node) RunSSHSetupNode() error {
	if err := h.RunOverSSH(
		"Setup Node",
		constants.SSHLongRunningScriptTimeout,
		"shell/setupNode.sh",
		scriptInputs{},
	); err != nil {
		return err
	}
	return nil
}

// RunSSHSetupDockerService runs script to setup docker compose service for CLI
func (h *Node) RunSSHSetupDockerService() error {
	if h.HasSystemDAvailable() {
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

// RunSSHRestartAvalanchego runs script to restart avalanchego
func (h *Node) RunSSHRestartAvalanchego() error {
	remoteComposeFile := utils.GetRemoteComposeFile()
	return h.RestartDockerComposeService(remoteComposeFile, constants.ServiceAvalanchego, constants.SSHLongRunningScriptTimeout)
}

// RunSSHStartAWMRelayerService runs script to start an AWM Relayer Service
func (h *Node) RunSSHStartAWMRelayerService() error {
	return h.StartDockerComposeService(utils.GetRemoteComposeFile(), constants.ServiceAWMRelayer, constants.SSHLongRunningScriptTimeout)
}

// RunSSHStopAWMRelayerService runs script to start an AWM Relayer Service
func (h *Node) RunSSHStopAWMRelayerService() error {
	return h.StopDockerComposeService(utils.GetRemoteComposeFile(), constants.ServiceAWMRelayer, constants.SSHLongRunningScriptTimeout)
}

// RunSSHUpgradeAvalanchego runs script to upgrade avalanchego
func (h *Node) RunSSHUpgradeAvalanchego(avalancheGoVersion string) error {
	withMonitoring, err := h.WasNodeSetupWithMonitoring()
	if err != nil {
		return err
	}

	if err := h.ComposeOverSSH("Compose Node",
		constants.SSHScriptTimeout,
		"templates/avalanchego.docker-compose.yml",
		dockerComposeInputs{
			AvalanchegoVersion: avalancheGoVersion,
			WithMonitoring:     withMonitoring,
			WithAvalanchego:    true,
			E2E:                utils.IsE2E(),
			E2EIP:              utils.E2EConvertIP(h.IP),
			E2ESuffix:          utils.E2ESuffix(h.IP),
		}); err != nil {
		return err
	}
	return h.RestartDockerCompose(constants.SSHLongRunningScriptTimeout)
}

// RunSSHStartAvalanchego runs script to start avalanchego
func (h *Node) RunSSHStartAvalanchego() error {
	return h.StartDockerComposeService(utils.GetRemoteComposeFile(), constants.ServiceAvalanchego, constants.SSHLongRunningScriptTimeout)
}

// RunSSHStopAvalanchego runs script to stop avalanchego
func (h *Node) RunSSHStopAvalanchego() error {
	return h.StopDockerComposeService(utils.GetRemoteComposeFile(), constants.ServiceAvalanchego, constants.SSHLongRunningScriptTimeout)
}

// RunSSHUpgradeSubnetEVM runs script to upgrade subnet evm
func (h *Node) RunSSHUpgradeSubnetEVM(subnetEVMBinaryPath string) error {
	if _, err := h.Commandf(nil, constants.SSHScriptTimeout, "cp -f subnet-evm %s", subnetEVMBinaryPath); err != nil {
		return err
	}
	return nil
}

func (h *Node) RunSSHSetupPrometheusConfig(avalancheGoPorts, machinePorts, loadTestPorts []string) error {
	for _, folder := range remoteconfig.PrometheusFoldersToCreate() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	cloudNodePrometheusConfigTemp := utils.GetRemoteComposeServicePath(constants.ServicePrometheus, "prometheus.yml")
	promConfig, err := os.CreateTemp("", constants.ServicePrometheus)
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

func (h *Node) RunSSHSetupLokiConfig(port int) error {
	for _, folder := range remoteconfig.LokiFoldersToCreate() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	cloudNodeLokiConfigTemp := utils.GetRemoteComposeServicePath(constants.ServiceLoki, "loki.yml")
	lokiConfig, err := os.CreateTemp("", constants.ServiceLoki)
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

func (h *Node) RunSSHSetupPromtailConfig(lokiIP string, lokiPort int, cloudID string, nodeID string, chainID string) error {
	for _, folder := range remoteconfig.PromtailFoldersToCreate() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	cloudNodePromtailConfigTemp := utils.GetRemoteComposeServicePath(constants.ServicePromtail, "promtail.yml")
	promtailConfig, err := os.CreateTemp("", constants.ServicePromtail)
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

func (h *Node) RunSSHUploadNodeAWMRelayerConfig(nodeInstanceDirPath string) error {
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
func (h *Node) RunSSHGetNewSubnetEVMRelease(subnetEVMReleaseURL, subnetEVMArchive string) error {
	return h.RunOverSSH(
		"Get Subnet EVM Release",
		constants.SSHScriptTimeout,
		"shell/getNewSubnetEVMRelease.sh",
		scriptInputs{SubnetEVMReleaseURL: subnetEVMReleaseURL, SubnetEVMArchive: subnetEVMArchive},
	)
}

// RunSSHUploadStakingFiles uploads staking files to a remote host via SSH.
func (h *Node) RunSSHUploadStakingFiles(keyPath string) error {
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

// RunSSHSetupMonitoringFolders sets up monitoring folders
func (h *Node) RunSSHSetupMonitoringFolders() error {
	for _, folder := range remoteconfig.RemoteFoldersToCreateMonitoring() {
		if err := h.MkdirAll(folder, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	return nil
}

// MonitorNodes links all the nodes specified with the monitoring node
// so that the monitoring host can start tracking the validator nodes metrics and collecting their
// logs
func (h *Node) MonitorNodes(ctx context.Context, targets []Node, chainID string) error {
	// nodesSet is a map with keys being format of targets.AWSProfile-targets.Region-targets.securityGroupName
	nodesSet := make(map[string]bool) // New empty set
	for _, node := range targets {
		nodeSetKeyName := fmt.Sprintf("%s|%s|%s", node.CloudConfig.AWSConfig.AWSProfile, node.CloudConfig.Region, node.CloudConfig.AWSConfig.AWSSecurityGroupName)
		nodesSet[nodeSetKeyName] = true
	}
	for nodeKey := range nodesSet {
		nodeInfo := strings.Split(nodeKey, "|")
		// Whitelist access to monitoring host IP address
		if err := awsAPI.WhitelistMonitoringAccess(ctx, nodeInfo[0], nodeInfo[1], nodeInfo[2], h.IP); err != nil {
			return fmt.Errorf("unable to whitelist monitoring access for node %s due to %s", h.NodeID, err.Error())
		}
	}
	// necessary checks
	if !isMonitoringNode(*h) {
		return fmt.Errorf("%s is not a monitoring node", h.NodeID)
	}
	for _, target := range targets {
		if isMonitoringNode(target) {
			return fmt.Errorf("target %s can't be a monitoring node", target.NodeID)
		}
	}
	if err := h.WaitForSSHShell(constants.SSHScriptTimeout); err != nil {
		return err
	}
	// setup monitoring for nodes
	remoteComposeFile := utils.GetRemoteComposeFile()
	wg := sync.WaitGroup{}
	wgResults := NodeResults{}
	for _, target := range targets {
		wg.Add(1)
		go func(nodeResults *NodeResults, target Node) {
			defer wg.Done()
			if err := target.RunSSHSetupPromtailConfig(h.IP, constants.AvalanchegoLokiPort, h.NodeID, h.NodeID, chainID); err != nil {
				nodeResults.AddResult(target.NodeID, nil, err)
				return
			}
			if err := target.RestartDockerComposeService(remoteComposeFile, constants.ServicePromtail, constants.SSHScriptTimeout); err != nil {
				nodeResults.AddResult(target.NodeID, nil, err)
				return
			}
		}(&wgResults, target)
	}
	wg.Wait()
	if wgResults.HasErrors() {
		return wgResults.Error()
	}
	// provide dashboards for targets
	tmpdir, err := os.MkdirTemp("", constants.ServiceGrafana)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	if err := monitoring.Setup(tmpdir); err != nil {
		return err
	}
	if err := h.RunSSHSetupMonitoringFolders(); err != nil {
		return err
	}
	if err := h.RunSSHCopyMonitoringDashboards(tmpdir); err != nil {
		return err
	}
	avalancheGoPorts, machinePorts, ltPorts := getPrometheusTargets(targets)
	h.Logger.Infof("avalancheGoPorts: %v, machinePorts: %v, ltPorts: %v", avalancheGoPorts, machinePorts, ltPorts)
	// reconfigure monitoring instance
	if err := h.RunSSHSetupLokiConfig(constants.AvalanchegoLokiPort); err != nil {
		return err
	}
	if err := h.RestartDockerComposeService(remoteComposeFile, constants.ServiceLoki, constants.SSHScriptTimeout); err != nil {
		return err
	}
	if err := h.RunSSHSetupPrometheusConfig(avalancheGoPorts, machinePorts, ltPorts); err != nil {
		return err
	}
	if err := h.RestartDockerComposeService(remoteComposeFile, constants.ServicePrometheus, constants.SSHScriptTimeout); err != nil {
		return err
	}

	return nil
}

// SyncSubnets reconfigures avalanchego to sync subnets
func (h *Node) SyncSubnets(subnetsToTrack []string) error {
	// necessary checks
	if !isAvalancheGoNode(*h) {
		return fmt.Errorf("%s is not a avalanchego node", h.NodeID)
	}
	withMonitoring, err := h.WasNodeSetupWithMonitoring()
	if err != nil {
		return err
	}
	if err := h.WaitForSSHShell(constants.SSHScriptTimeout); err != nil {
		return err
	}
	avagoVersion, err := h.GetDockerImageVersion(constants.AvalancheGoDockerImage, constants.SSHScriptTimeout)
	if err != nil {
		return err
	}
	networkName, err := h.GetAvalancheGoNetworkName()
	if err != nil {
		return err
	}
	if err := h.ComposeSSHSetupNode(networkName, subnetsToTrack, avagoVersion, withMonitoring); err != nil {
		return err
	}
	if err := h.RestartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}

	return nil
}

func (h *Node) RunSSHCopyMonitoringDashboards(monitoringDashboardPath string) error {
	// TODO: download dashboards from github instead
	remoteDashboardsPath := utils.GetRemoteComposeServicePath("grafana", "dashboards")
	if !utils.DirectoryExists(monitoringDashboardPath) {
		return fmt.Errorf("%s does not exist", monitoringDashboardPath)
	}
	if err := h.MkdirAll(remoteDashboardsPath, constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	monitoringDashboardPath = filepath.Join(monitoringDashboardPath, constants.DashboardsDir)
	dashboards, err := os.ReadDir(monitoringDashboardPath)
	if err != nil {
		return err
	}
	for _, dashboard := range dashboards {
		if err := h.Upload(
			filepath.Join(monitoringDashboardPath, dashboard.Name()),
			filepath.Join(remoteDashboardsPath, dashboard.Name()),
			constants.SSHFileOpsTimeout,
		); err != nil {
			return err
		}
	}
	if composeFileExists(*h) {
		return h.RestartDockerComposeService(utils.GetRemoteComposeFile(), constants.ServiceGrafana, constants.SSHScriptTimeout)
	}
	return nil
}
