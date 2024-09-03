// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package interchainmessenger

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ethereum/go-ethereum/common"
)

const (
	messengerContractAddressURLFmt = "TeleporterMessenger_Contract_Address_%s.txt"
	messengerDeployerAddressURLFmt = "TeleporterMessenger_Deployer_Address_%s.txt"
	messengerDeployerTxURLFmt      = "TeleporterMessenger_Deployment_Transaction_%s.txt"
	registryBytecodeURLFmt         = "TeleporterRegistry_Bytecode_%s.txt"
)

var (
	// 10 TOKENS
	messengerDeployerRequiredBalance = big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10))
	// 600 TOKENS
	ICMRequiredBalance = big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(600))
)

func GetLatestVersion() (string, error) {
	return utils.GetLatestGithubReleaseVersion(constants.AvaLabsOrg, constants.ICMRepoName, "")
}

func getURLs(version string) (string, string, string, string) {
	messengerContractAddressURL := utils.GetGithubReleaseAssetURL(
		constants.AvaLabsOrg,
		constants.ICMRepoName,
		version,
		fmt.Sprintf(messengerContractAddressURLFmt, version),
	)
	messengerDeployerAddressURL := utils.GetGithubReleaseAssetURL(
		constants.AvaLabsOrg,
		constants.ICMRepoName,
		version,
		fmt.Sprintf(messengerDeployerAddressURLFmt, version),
	)
	messengerDeployerTxURL := utils.GetGithubReleaseAssetURL(
		constants.AvaLabsOrg,
		constants.ICMRepoName,
		version,
		fmt.Sprintf(messengerDeployerTxURLFmt, version),
	)
	registryBytecodeURL := utils.GetGithubReleaseAssetURL(
		constants.AvaLabsOrg,
		constants.ICMRepoName,
		version,
		fmt.Sprintf(registryBytecodeURLFmt, version),
	)
	return messengerContractAddressURL, messengerDeployerAddressURL, messengerDeployerTxURL, registryBytecodeURL
}

type Deployer struct {
	messengerContractAddress []byte
	messengerDeployerAddress []byte
	messengerDeployerTx      []byte
	registryBytecode         []byte
}

func (t *Deployer) CheckAssets() error {
	if len(t.messengerContractAddress) == 0 || len(t.messengerDeployerAddress) == 0 || len(t.messengerDeployerTx) == 0 || len(t.registryBytecode) == 0 {
		return fmt.Errorf("interchain messaging assets has not been initialized")
	}
	return nil
}

func (t *Deployer) GetAssets() ([]byte, []byte, []byte, []byte) {
	return t.messengerContractAddress, t.messengerDeployerAddress, t.messengerDeployerTx, t.registryBytecode
}

func (t *Deployer) SetAssets(
	messengerContractAddress []byte,
	messengerDeployerAddress []byte,
	messengerDeployerTx []byte,
	registryBytecode []byte,
) {
	t.messengerContractAddress = messengerContractAddress
	t.messengerDeployerAddress = messengerDeployerAddress
	t.messengerDeployerTx = messengerDeployerTx
	t.registryBytecode = registryBytecode
}

func (t *Deployer) LoadAssets(
	messengerContractAddressPath string,
	messengerDeployerAddressPath string,
	messengerDeployerTxPath string,
	registryBytecodePath string,
) error {
	var err error
	if messengerContractAddressPath != "" {
		if t.messengerContractAddress, err = os.ReadFile(messengerContractAddressPath); err != nil {
			return err
		}
	}
	if messengerDeployerAddressPath != "" {
		if t.messengerDeployerAddress, err = os.ReadFile(messengerDeployerAddressPath); err != nil {
			return err
		}
	}
	if messengerDeployerTxPath != "" {
		if t.messengerDeployerTx, err = os.ReadFile(messengerDeployerTxPath); err != nil {
			return err
		}
	}
	if registryBytecodePath != "" {
		if t.registryBytecode, err = os.ReadFile(registryBytecodePath); err != nil {
			return err
		}
	}
	return nil
}

func (t *Deployer) DownloadAssets(version string) error {
	var err error
	messengerContractAddressURL, messengerDeployerAddressURL, messengerDeployerTxURL, registryBytecodeURL := getURLs(version)
	if t.messengerContractAddress, err = utils.HTTPGet(messengerContractAddressURL, ""); err != nil {
		return err
	}
	if t.messengerDeployerAddress, err = utils.HTTPGet(messengerDeployerAddressURL, ""); err != nil {
		return err
	}
	if t.messengerDeployerTx, err = utils.HTTPGet(messengerDeployerTxURL, ""); err != nil {
		return err
	}
	if t.registryBytecode, err = utils.HTTPGet(registryBytecodeURL, ""); err != nil {
		return err
	}
	return nil
}

// Deploys messenger and registry
// If messenger is already deployed, will avoid deploying registry,
// unless [forceRegistryDeploy] is set
func (t *Deployer) Deploy(
	rpcURL string,
	privateKey string,
	forceRegistryDeploy bool,
) (bool, string, string, error) {
	var registryAddress string
	messengerAlreadyDeployed, messengerAddress, err := t.DeployMessenger(
		rpcURL,
		privateKey,
	)
	if !messengerAlreadyDeployed || forceRegistryDeploy {
		registryAddress, err = t.DeployRegistry(rpcURL, privateKey)
	}
	return messengerAlreadyDeployed, registryAddress, messengerAddress, err
}

func (t *Deployer) DeployMessenger(
	rpcURL string,
	privateKey string,
) (bool, string, error) {
	if err := t.CheckAssets(); err != nil {
		return false, "", err
	}
	client, err := evm.GetClient(rpcURL)
	if err != nil {
		return false, "", err
	}
	if messengerAlreadyDeployed, err := evm.ContractAlreadyDeployed(client, string(t.messengerContractAddress)); err != nil {
		return false, "", fmt.Errorf("failure making a request to %s: %w", rpcURL, err)
	} else if messengerAlreadyDeployed {
		return true, string(t.messengerContractAddress), nil
	}
	if err := evm.SetMinBalance(
		client,
		privateKey,
		string(t.messengerDeployerAddress),
		messengerDeployerRequiredBalance,
	); err != nil {
		return false, "", err
	}
	if err := evm.IssueTx(client, string(t.messengerDeployerTx)); err != nil {
		return false, "", err
	}
	return false, string(t.messengerContractAddress), nil
}

func (t *Deployer) DeployRegistry(
	rpcURL string,
	privateKey string,
) (string, error) {
	if err := t.CheckAssets(); err != nil {
		return "", err
	}
	messengerContractAddress := common.HexToAddress(string(t.messengerContractAddress))
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
	registryAddress, err := evm.DeployContract(
		rpcURL,
		privateKey,
		t.registryBytecode,
		"([(uint256, address)])",
		constructorInput,
	)
	if err != nil {
		return "", err
	}
	return registryAddress.Hex(), nil
}
