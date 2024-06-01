// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package models

type Sidecar struct {
	Name                string
	VM                  VMType
	VMVersion           string
	RPCVersion          int
	Subnet              string
	TokenName           string
	TokenSymbol         string
	ChainID             string
	Version             string
	Networks            map[string]NetworkData
	ElasticSubnet       map[string]ElasticSubnet
	ImportedFromAPM     bool
	ImportedVMID        string
	CustomVMRepoURL     string
	CustomVMBranch      string
	CustomVMBuildScript string
	// Teleporter related
	TeleporterReady   bool
	TeleporterKey     string
	TeleporterVersion string
	RunRelayer        bool
	// SubnetEVM based VM's only
	SubnetEVMMainnetChainID uint
}

type Exportable struct {
	Sidecar         Sidecar
	Genesis         []byte
	ChainConfig     []byte
	SubnetConfig    []byte
	NetworkUpgrades []byte
	NodeConfig      []byte
}
