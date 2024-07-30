// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	remoteconfig "github.com/ava-labs/avalanche-tooling-sdk-go/node/config"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// PrepareAvalanchegoConfig creates the config files for the AvalancheGo
// networkID is the ID of the network to be used
// trackSubnets is the list of subnets to track
func (h *Node) RunSSHRenderAvalancheNodeConfig(networkID string, trackSubnets []string) error {
	avagoConf := remoteconfig.PrepareAvalancheConfig(h.IP, networkID, trackSubnets)

	nodeConf, err := remoteconfig.RenderAvalancheNodeConfig(avagoConf)
	if err != nil {
		return err
	}
	// preserve remote configuration if it exists
	if nodeConfigFileExists(*h) {
		// make sure that bootsrap configuration is preserved
		if genesisFileExists(*h) {
			avagoConf.GenesisPath = remoteconfig.GetRemoteAvalancheGenesis()
		}
		remoteAvagoConf, err := h.GetAvalancheGoConfigData()
		if err != nil {
			return err
		}
		// ignore errors if bootstrap configuration is not present - it's fine
		bootstrapIDs, _ := utils.StringValue(remoteAvagoConf, "bootstrap-ids")
		bootstrapIPs, _ := utils.StringValue(remoteAvagoConf, "bootstrap-ips")

		avagoConf.BootstrapIDs = bootstrapIDs
		avagoConf.BootstrapIPs = bootstrapIPs
	}
	// configuration is ready to be uploaded
	if err := h.UploadBytes(nodeConf, remoteconfig.GetRemoteAvalancheNodeConfig(), constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	cChainConf, err := remoteconfig.RenderAvalancheCChainConfig(avagoConf)
	if err != nil {
		return err
	}
	if err := h.UploadBytes(cChainConf, remoteconfig.GetRemoteAvalancheCChainConfig(), constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	return nil
}

func prepareGrafanaConfig() (string, string, string, string, error) {
	grafanaDataSource, err := remoteconfig.RenderGrafanaLokiDataSourceConfig()
	if err != nil {
		return "", "", "", "", err
	}
	grafanaDataSourceFile, err := os.CreateTemp("", "avalanchecli-grafana-datasource-*.yml")
	if err != nil {
		return "", "", "", "", err
	}
	if err := os.WriteFile(grafanaDataSourceFile.Name(), grafanaDataSource, constants.WriteReadUserOnlyPerms); err != nil {
		return "", "", "", "", err
	}

	grafanaPromDataSource, err := remoteconfig.RenderGrafanaPrometheusDataSourceConfigg()
	if err != nil {
		return "", "", "", "", err
	}
	grafanaPromDataSourceFile, err := os.CreateTemp("", "avalanchecli-grafana-prom-datasource-*.yml")
	if err != nil {
		return "", "", "", "", err
	}
	if err := os.WriteFile(grafanaPromDataSourceFile.Name(), grafanaPromDataSource, constants.WriteReadUserOnlyPerms); err != nil {
		return "", "", "", "", err
	}

	grafanaDashboards, err := remoteconfig.RenderGrafanaDashboardConfig()
	if err != nil {
		return "", "", "", "", err
	}
	grafanaDashboardsFile, err := os.CreateTemp("", "avalanchecli-grafana-dashboards-*.yml")
	if err != nil {
		return "", "", "", "", err
	}
	if err := os.WriteFile(grafanaDashboardsFile.Name(), grafanaDashboards, constants.WriteReadUserOnlyPerms); err != nil {
		return "", "", "", "", err
	}

	grafanaConfig, err := remoteconfig.RenderGrafanaConfig()
	if err != nil {
		return "", "", "", "", err
	}
	grafanaConfigFile, err := os.CreateTemp("", "avalanchecli-grafana-config-*.ini")
	if err != nil {
		return "", "", "", "", err
	}
	if err := os.WriteFile(grafanaConfigFile.Name(), grafanaConfig, constants.WriteReadUserOnlyPerms); err != nil {
		return "", "", "", "", err
	}
	return grafanaConfigFile.Name(), grafanaDashboardsFile.Name(), grafanaDataSourceFile.Name(), grafanaPromDataSourceFile.Name(), nil
}
