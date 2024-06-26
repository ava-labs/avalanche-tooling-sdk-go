// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package services

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

func RenderGrafanaLokiDataSourceConfig() ([]byte, error) {
	return templates.ReadFile("templates/grafana-loki-datasource.yaml")
}

func RenderGrafanaPrometheusDataSourceConfigg() ([]byte, error) {
	return templates.ReadFile("templates/grafana-prometheus-datasource.yaml")
}

func RenderGrafanaConfig() ([]byte, error) {
	return templates.ReadFile("templates/grafana.ini")
}

func RenderGrafanaDashboardConfig() ([]byte, error) {
	return templates.ReadFile("templates/grafana-dashboards.yaml")
}

func GrafanaFoldersToCreate() []string {
	return []string{
		utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "data"),
		utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "dashboards"),
		utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "provisioning", "datasources"),
		utils.GetRemoteComposeServicePath(constants.ServiceGrafana, "provisioning", "dashboards"),
	}
}
