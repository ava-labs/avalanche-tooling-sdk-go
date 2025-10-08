// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package icm provides utilities for deploying and interacting with
// Interchain Messaging (ICM) contracts on Avalanche L1s.
//
// ICM (formerly known as Teleporter) enables asynchronous cross-chain
// communication between EVM-based Avalanche L1s using the Avalanche Warp
// Messaging protocol.
package icm

import (
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/libevm/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm/contract"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

//go:embed smart_contracts/messenger_deployment_transaction_v1.0.0.txt
var defaultMessengerDeployerTx []byte

//go:embed smart_contracts/registry_bytecode_v1.0.0.txt
var defaultRegistryBytecode []byte

const (
	// DefaultVersion is the default ICM contracts release version.
	DefaultVersion = "v1.0.0"

	// DefaultMessengerContractAddress is the canonical address of the TeleporterMessenger contract.
	// This address is deterministic across all chains when deployed using Nick's method.
	DefaultMessengerContractAddress = "0x253b2784c75e510dD0fF1da844684a1aC0aa5fcf"

	// DefaultMessengerDeployerAddress is the deployer address for the messenger contract.
	DefaultMessengerDeployerAddress = "0x618FEdD9A45a8C456812ecAAE70C671c6249DfaC"

	releaseURL                     = "https://github.com/ava-labs/icm-contracts/releases/download/%s/"
	messengerContractAddressURLFmt = releaseURL + "/TeleporterMessenger_Contract_Address_%s.txt"
	messengerDeployerAddressURLFmt = releaseURL + "/TeleporterMessenger_Deployer_Address_%s.txt"
	messengerDeployerTxURLFmt      = releaseURL + "/TeleporterMessenger_Deployment_Transaction_%s.txt"
	registryBytecodeURLFmt         = releaseURL + "/TeleporterRegistry_Bytecode_%s.txt"
)

var (
	messengerDeployerRequiredBalance = big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10)) // 10 AVAX

	// ErrMessengerAlreadyDeployed is returned when the messenger contract is already deployed on the chain.
	ErrMessengerAlreadyDeployed = errors.New("messenger already deployed")
)

// GetReleaseURLs returns the GitHub release URLs for ICM contract deployment assets.
// Returns URLs for: messenger contract address, messenger deployer address,
// messenger deployment transaction, and registry bytecode.
func GetReleaseURLs(version string) (string, string, string, string) {
	messengerContractAddressURL := fmt.Sprintf(
		messengerContractAddressURLFmt,
		version,
		version,
	)
	messengerDeployerAddressURL := fmt.Sprintf(
		messengerDeployerAddressURLFmt,
		version,
		version,
	)
	messengerDeployerTxURL := fmt.Sprintf(
		messengerDeployerTxURLFmt,
		version,
		version,
	)
	registryBydecodeURL := fmt.Sprintf(
		registryBytecodeURLFmt,
		version,
		version,
	)
	return messengerContractAddressURL, messengerDeployerAddressURL, messengerDeployerTxURL, registryBydecodeURL
}

// Deployer manages deployment of ICM contracts (TeleporterMessenger and TeleporterRegistry).
// Use LoadDefault, LoadFromFiles, or LoadFromRelease to initialize deployment assets before deploying.
type Deployer struct {
	messengerContractAddress string
	messengerDeployerAddress string
	messengerDeployerTx      []byte
	registryBydecode         []byte
}

// Get returns the deployment assets: messenger contract address, messenger deployer address,
// messenger deployment transaction, and registry bytecode.
func (t *Deployer) Get() (string, string, []byte, []byte) {
	return t.messengerContractAddress, t.messengerDeployerAddress, t.messengerDeployerTx, t.registryBydecode
}

// Validate checks that all deployment assets have been loaded.
func (t *Deployer) Validate() error {
	if t.messengerContractAddress == "" || t.messengerDeployerAddress == "" || len(t.messengerDeployerTx) == 0 || len(t.registryBydecode) == 0 {
		return fmt.Errorf("icm assets has not been initialized")
	}
	return nil
}

// LoadFromFiles loads deployment assets from local file paths.
func (t *Deployer) LoadFromFiles(
	messengerContractAddressPath string,
	messengerDeployerAddressPath string,
	messengerDeployerTxPath string,
	registryBydecodePath string,
) error {
	messengerContractAddressBytes, err := os.ReadFile(messengerContractAddressPath)
	if err != nil {
		return err
	}
	messengerDeployerAddressBytes, err := os.ReadFile(messengerDeployerAddressPath)
	if err != nil {
		return err
	}
	messengerDeployerTx, err := os.ReadFile(messengerDeployerTxPath)
	if err != nil {
		return err
	}
	registryBydecode, err := os.ReadFile(registryBydecodePath)
	if err != nil {
		return err
	}
	t.Set(
		string(messengerContractAddressBytes),
		string(messengerDeployerAddressBytes),
		messengerDeployerTx,
		registryBydecode,
	)
	return nil
}

// Set sets the deployment assets directly.
func (t *Deployer) Set(
	messengerContractAddress string,
	messengerDeployerAddress string,
	messengerDeployerTx []byte,
	registryBydecode []byte,
) {
	t.messengerContractAddress = messengerContractAddress
	t.messengerDeployerAddress = messengerDeployerAddress
	t.messengerDeployerTx = messengerDeployerTx
	t.registryBydecode = registryBydecode
}

// LoadDefault loads the default embedded deployment assets for ICM v1.0.0.
func (t *Deployer) LoadDefault() {
	t.Set(
		DefaultMessengerContractAddress,
		DefaultMessengerDeployerAddress,
		defaultMessengerDeployerTx,
		defaultRegistryBytecode,
	)
}

// LoadFromRelease downloads deployment assets from a GitHub release.
// The token parameter is optional and can be empty for public releases.
func (t *Deployer) LoadFromRelease(
	version string,
	token string,
) error {
	messengerContractAddressURL, messengerDeployerAddressURL, messengerDeployerTxURL, registryBydecodeURL := GetReleaseURLs(
		version,
	)
	downloader := utils.NewDownloader()
	messengerContractAddressBytes, err := downloader.Download(messengerContractAddressURL, token)
	if err != nil {
		return err
	}
	messengerDeployerAddressBytes, err := downloader.Download(messengerDeployerAddressURL, token)
	if err != nil {
		return err
	}
	messengerDeployerTxBytes, err := downloader.Download(messengerDeployerTxURL, token)
	if err != nil {
		return err
	}
	registryBytecodeBytes, err := downloader.Download(registryBydecodeURL, token)
	if err != nil {
		return err
	}
	t.Set(
		string(messengerContractAddressBytes),
		string(messengerDeployerAddressBytes),
		messengerDeployerTxBytes,
		registryBytecodeBytes,
	)
	return nil
}

// Deploy deploys both the TeleporterMessenger and TeleporterRegistry contracts.
// Returns the messenger address, registry address, and any error encountered.
// If the messenger is already deployed, returns ErrMessengerAlreadyDeployed.
func (t *Deployer) Deploy(
	rpcURL string,
	privateKey string,
) (string, string, error) {
	messengerAddress, err := t.DeployMessenger(rpcURL, privateKey)
	if err != nil {
		return messengerAddress, "", err
	}
	registryAddress, err := t.DeployRegistry(rpcURL, privateKey)
	if err != nil {
		return messengerAddress, "", err
	}
	return messengerAddress, registryAddress, nil
}

// DeployMessenger deploys the TeleporterMessenger contract using Nick's method.
// It automatically funds the deployer address if needed (minimum 10 AVAX).
// Returns the messenger contract address and ErrMessengerAlreadyDeployed if already deployed.
func (t *Deployer) DeployMessenger(
	rpcURL string,
	privateKey string,
) (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	// check if contract is already deployed
	client, err := evm.GetClient(rpcURL)
	if err != nil {
		return "", err
	}
	messengerAlreadyDeployed, err := client.ContractAlreadyDeployed(t.messengerContractAddress)
	if err != nil {
		return "", fmt.Errorf("failure making a request to %s: %w", rpcURL, err)
	}
	if messengerAlreadyDeployed {
		return t.messengerContractAddress, ErrMessengerAlreadyDeployed
	}
	messengerDeployerBalance, err := client.GetAddressBalance(
		t.messengerDeployerAddress,
	)
	if err != nil {
		return "", err
	}
	if messengerDeployerBalance.Cmp(messengerDeployerRequiredBalance) < 0 {
		toFund := big.NewInt(0).
			Sub(messengerDeployerRequiredBalance, messengerDeployerBalance)
		if _, err := client.FundAddress(
			privateKey,
			t.messengerDeployerAddress,
			toFund,
		); err != nil {
			return "", err
		}
	}
	if err := client.IssueTx(string(t.messengerDeployerTx)); err != nil {
		return "", err
	}
	return t.messengerContractAddress, nil
}

// DeployRegistry deploys the TeleporterRegistry contract.
// The registry is initialized with the messenger contract address at version 1.
func (t *Deployer) DeployRegistry(
	rpcURL string,
	privateKey string,
) (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	messengerContractAddress := common.HexToAddress(t.messengerContractAddress)
	type ProtocolRegistryEntry struct {
		Version         *big.Int
		ProtocolAddress common.Address
	}
	constructorInput := []ProtocolRegistryEntry{
		{
			Version:         big.NewInt(1),
			ProtocolAddress: messengerContractAddress,
		},
	}
	registryAddress, _, _, err := contract.DeployContract(
		rpcURL,
		privateKey,
		t.registryBydecode,
		"([(uint256, address)])",
		constructorInput,
	)
	if err != nil {
		return "", err
	}
	return registryAddress.Hex(), nil
}
