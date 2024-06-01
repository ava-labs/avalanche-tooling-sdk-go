// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"encoding/json"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/models"
	"os"
)

// ExportSubnet exports subnet into a file that can be imported into Avalanche-CLI
func (subnet Subnet) ExportSubnet(exportOutput string) error {
	var err error
	if exportOutput == "" {
		return fmt.Errorf("export file path is missing")
	}
	//subnetName := subnet.Name
	//
	//sc, err := app.LoadSidecar(subnetName)
	//if err != nil {
	//	return err
	//}

	//gen, err := subnet.Genesis
	//if err != nil {
	//	return err
	//}

	var nodeConfig, chainConfig, subnetConfig, networkUpgrades []byte

	//if app.AvagoNodeConfigExists(subnetName) {
	//	nodeConfig, err = app.LoadRawAvagoNodeConfig(subnetName)
	//	if err != nil {
	//		return err
	//	}
	//}
	//if app.ChainConfigExists(subnetName) {
	//	chainConfig, err = app.LoadRawChainConfig(subnetName)
	//	if err != nil {
	//		return err
	//	}
	//}
	//if app.AvagoSubnetConfigExists(subnetName) {
	//	subnetConfig, err = app.LoadRawAvagoSubnetConfig(subnetName)
	//	if err != nil {
	//		return err
	//	}
	//}
	//if app.NetworkUpgradeExists(subnetName) {
	//	networkUpgrades, err = app.LoadRawNetworkUpgrades(subnetName)
	//	if err != nil {
	//		return err
	//	}
	//}
	sidecar := models.Sidecar{}
	exportData := models.Exportable{
		Sidecar:         sc,
		Genesis:         gen,
		NodeConfig:      nodeConfig,
		ChainConfig:     chainConfig,
		SubnetConfig:    subnetConfig,
		NetworkUpgrades: networkUpgrades,
	}

	exportBytes, err := json.Marshal(exportData)
	if err != nil {
		return err
	}
	return os.WriteFile(exportOutput, exportBytes, constants.WriteReadReadPerms)
}
