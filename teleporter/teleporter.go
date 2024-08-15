// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package teleporter

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ava-labs/avalanche-cli/pkg/application"
	"github.com/ava-labs/avalanche-cli/pkg/binutils"
	"github.com/ava-labs/avalanche-cli/pkg/constants"
	"github.com/ava-labs/avalanche-cli/pkg/contract"
	"github.com/ava-labs/avalanche-cli/pkg/evm"
	"github.com/ava-labs/avalanche-cli/pkg/key"
	"github.com/ava-labs/avalanche-cli/pkg/models"
	"github.com/ava-labs/avalanche-cli/pkg/ux"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ethereum/go-ethereum/common"
)

const (
	releaseURL                     = "https://github.com/ava-labs/teleporter/releases/download/%s/"
	messengerContractAddressURLFmt = releaseURL + "/TeleporterMessenger_Contract_Address_%s.txt"
	messengerDeployerAddressURLFmt = releaseURL + "/TeleporterMessenger_Deployer_Address_%s.txt"
	messengerDeployerTxURLFmt      = releaseURL + "/TeleporterMessenger_Deployment_Transaction_%s.txt"
	registryBytecodeURLFmt         = releaseURL + "/TeleporterRegistry_Bytecode_%s.txt"
)

var (
	// 10 AVAX
	messengerDeployerRequiredBalance = big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10))
	// 600 AVAX
	InterchainMessagingPrefundedAddressBalance = big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(600))
)

func getTeleporterURLs(version string) (string, string, string, string) {
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

type Deployer struct {
	messengerContractAddress string
	messengerDeployerAddress string
	messengerDeployerTx      string
	registryBydecode         string
}

func (t *Deployer) GetAssets(
	teleporterInstallDir string,
	version string,
) (string, string, string, string, error) {
	if err := t.DownloadAssets(teleporterInstallDir, version); err != nil {
		return "", "", "", "", err
	}
	return t.messengerContractAddress, t.messengerDeployerAddress, t.messengerDeployerTx, t.registryBydecode, nil
}

func (t *Deployer) CheckAssets() error {
	if t.messengerContractAddress == "" || t.messengerDeployerAddress == "" || t.messengerDeployerTx == "" || t.registryBydecode == "" {
		return fmt.Errorf("teleporter assets has not been initialized")
	}
	return nil
}

func (t *Deployer) SetAssetsFromPaths(
	messengerContractAddressPath string,
	messengerDeployerAddressPath string,
	messengerDeployerTxPath string,
	registryBydecodePath string,
) error {
	if messengerContractAddressPath != "" {
		if bs, err := os.ReadFile(messengerContractAddressPath); err != nil {
			return err
		} else {
			t.messengerContractAddress = string(bs)
		}
	}
	if messengerDeployerAddressPath != "" {
		if bs, err := os.ReadFile(messengerDeployerAddressPath); err != nil {
			return err
		} else {
			t.messengerDeployerAddress = string(bs)
		}
	}
	if messengerDeployerTxPath != "" {
		if bs, err := os.ReadFile(messengerDeployerTxPath); err != nil {
			return err
		} else {
			t.messengerDeployerTx = string(bs)
		}
	}
	if registryBydecodePath != "" {
		if bs, err := os.ReadFile(registryBydecodePath); err != nil {
			return err
		} else {
			t.registryBydecode = string(bs)
		}
	}
	return nil
}

func (t *Deployer) SetAssets(
	messengerContractAddress string,
	messengerDeployerAddress string,
	messengerDeployerTx string,
	registryBydecode string,
) {
	if messengerContractAddress != "" {
		t.messengerContractAddress = messengerContractAddress
	}
	if messengerDeployerAddress != "" {
		t.messengerDeployerAddress = messengerDeployerAddress
	}
	if messengerDeployerTx != "" {
		t.messengerDeployerTx = messengerDeployerTx
	}
	if registryBydecode != "" {
		t.registryBydecode = registryBydecode
	}
}

func (t *Deployer) DownloadAssets(version string) error {
	messengerContractAddressURL, messengerDeployerAddressURL, messengerDeployerTxURL, registryBydecodeURL := getTeleporterURLs(version)
	if t.messengerContractAddress == "" {
		// get target teleporter messenger contract address
		messengerContractAddressBytes, err := utils.HTTPGet(messengerContractAddressURL, "")
		if err != nil {
			return err
		}
		t.messengerContractAddress = string(messengerContractAddressBytes)
	}
	if t.messengerDeployerAddress == "" {
		// get teleporter deployer address
		messengerDeployerAddressBytes, err := utils.HTTPGet(messengerDeployerAddressURL, "")
		if err != nil {
			return err
		}
		t.messengerDeployerAddress = string(messengerDeployerAddressBytes)
	}
	if t.messengerDeployerTx == "" {
		messengerDeployerTxBytes, err := utils.HTTPGet(messengerDeployerTxURL, "")
		if err != nil {
			return err
		}
		t.messengerDeployerTx = string(messengerDeployerTxBytes)
	}
	if t.registryBydecode == "" {
		registryBytecodeBytes, err := utils.HTTPGet(registryBydecodeURL, "")
		if err != nil {
			return err
		}
		t.registryBydecode = string(registryBytecodeBytes)
	}
	return nil
}

func (t *Deployer) Deploy(
	subnetName string,
	rpcURL string,
	privateKey string,
	deployMessenger bool,
	deployRegistry bool,
) (bool, string, string, error) {
	var (
		messengerAddress string
		registryAddress  string
		alreadyDeployed  bool
		err              error
	)
	if deployMessenger {
		alreadyDeployed, messengerAddress, err = t.DeployMessenger(
			subnetName,
			rpcURL,
			privateKey,
		)
	}
	if err == nil && deployRegistry {
		if !deployMessenger || !alreadyDeployed {
			registryAddress, err = t.DeployRegistry(subnetName, rpcURL, privateKey)
		}
	}
	return alreadyDeployed, messengerAddress, registryAddress, err
}

func (t *Deployer) DeployMessenger(
	subnetName string,
	rpcURL string,
	privateKey string,
) (bool, string, error) {
	if err := t.CheckAssets(); err != nil {
		return false, "", err
	}
	// check if contract is already deployed
	client, err := evm.GetClient(rpcURL)
	if err != nil {
		return false, "", err
	}
	if messengerAlreadyDeployed, err := evm.ContractAlreadyDeployed(client, t.messengerContractAddress); err != nil {
		return false, "", fmt.Errorf("failure making a request to %s: %w", rpcURL, err)
	} else if messengerAlreadyDeployed {
		ux.Logger.PrintToUser("Teleporter Messenger has already been deployed to %s", subnetName)
		return true, t.messengerContractAddress, nil
	}
	// get teleporter deployer balance
	messengerDeployerBalance, err := evm.GetAddressBalance(
		client,
		t.messengerDeployerAddress,
	)
	if err != nil {
		return false, "", err
	}
	if messengerDeployerBalance.Cmp(messengerDeployerRequiredBalance) < 0 {
		toFund := big.NewInt(0).
			Sub(messengerDeployerRequiredBalance, messengerDeployerBalance)
		if err := evm.FundAddress(
			client,
			privateKey,
			t.messengerDeployerAddress,
			toFund,
		); err != nil {
			return false, "", err
		}
	}
	if err := evm.IssueTx(client, t.messengerDeployerTx); err != nil {
		return false, "", err
	}
	ux.Logger.PrintToUser(
		"Teleporter Messenger successfully deployed to %s (%s)",
		subnetName,
		t.messengerContractAddress,
	)
	return false, t.messengerContractAddress, nil
}

func (t *Deployer) DeployRegistry(
	subnetName string,
	rpcURL string,
	privateKey string,
) (string, error) {
	if err := t.CheckAssets(); err != nil {
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
	registryAddress, err := contract.DeployContract(
		rpcURL,
		privateKey,
		[]byte(t.registryBydecode),
		"([(uint256, address)])",
		constructorInput,
	)
	if err != nil {
		return "", err
	}
	ux.Logger.PrintToUser(
		"Teleporter Registry successfully deployed to %s (%s)",
		subnetName,
		registryAddress,
	)
	return registryAddress.Hex(), nil
}

func getPrivateKey(
	app *application.Avalanche,
	network models.Network,
	keyName string,
) (string, error) {
	var (
		err error
		k   *key.SoftKey
	)
	if keyName == "" {
		if k, err = key.LoadEwoq(network.ID); err != nil {
			return "", err
		}
	} else {
		k, err = key.LoadSoft(network.ID, app.GetKeyPath(keyName))
		if err != nil {
			return "", err
		}
	}
	return k.PrivKeyHex(), nil
}

func SetProposerVM(
	app *application.Avalanche,
	network models.Network,
	blockchainID string,
	fundedKeyName string,
) error {
	privKeyStr, err := getPrivateKey(app, network, fundedKeyName)
	if err != nil {
		return err
	}
	wsEndpoint := network.BlockchainWSEndpoint(blockchainID)
	return evm.SetupProposerVM(wsEndpoint, privKeyStr)
}

func DeployAndFundRelayer(
	app *application.Avalanche,
	td *Deployer,
	network models.Network,
	subnetName string,
	blockchainID string,
	fundedKeyName string,
) (bool, string, string, error) {
	privKeyStr, err := getPrivateKey(app, network, fundedKeyName)
	if err != nil {
		return false, "", "", err
	}
	endpoint := network.BlockchainEndpoint(blockchainID)
	alreadyDeployed, messengerAddress, registryAddress, err := td.Deploy(
		subnetName,
		endpoint,
		privKeyStr,
		true,
		true,
	)
	if err != nil {
		return false, "", "", err
	}
	if !alreadyDeployed {
		// get relayer address
		relayerAddress, _, err := GetRelayerKeyInfo(app.GetKeyPath(constants.AWMRelayerKeyName))
		if err != nil {
			return false, "", "", err
		}
		// fund relayer
		if err := FundRelayer(
			endpoint,
			privKeyStr,
			relayerAddress,
		); err != nil {
			return false, "", "", err
		}
	}
	return alreadyDeployed, messengerAddress, registryAddress, err
}

func getTeleporterKeyInfo(
	app *application.Avalanche,
	keyName string,
) (string, string, *big.Int, error) {
	k, err := key.LoadSoftOrCreate(models.NewLocalNetwork().ID, app.GetKeyPath(keyName))
	if err != nil {
		return "", "", nil, err
	}
	return k.C(), k.PrivKeyHex(), InterchainMessagingPrefundedAddressBalance, nil
}

type Info struct {
	Version                  string
	FundedAddress            string
	FundedBalance            *big.Int
	MessengerDeployerAddress string
	RelayerAddress           string
}

func GetInfo(
	app *application.Avalanche,
) (*Info, error) {
	var err error
	ti := Info{}
	ti.FundedAddress, _, ti.FundedBalance, err = getTeleporterKeyInfo(
		app,
		constants.TeleporterKeyName,
	)
	if err != nil {
		return nil, err
	}
	ti.Version, err = app.Downloader.GetLatestReleaseVersion(
		binutils.GetGithubLatestReleaseURL(constants.AvaLabsOrg, constants.TeleporterRepoName),
	)
	if err != nil {
		return nil, err
	}
	deployer := Deployer{}
	_, ti.MessengerDeployerAddress, _, _, err = deployer.GetAssets(
		app.GetTeleporterBinDir(),
		ti.Version,
	)
	if err != nil {
		return nil, err
	}
	ti.RelayerAddress, _, err = GetRelayerKeyInfo(app.GetKeyPath(constants.AWMRelayerKeyName))
	if err != nil {
		return nil, err
	}
	return &ti, nil
}
