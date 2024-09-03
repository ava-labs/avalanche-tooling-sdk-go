// Copyright (C) 2022, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.
package localrelayer

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/install"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/process"
)

const (
	localRelayerSetupTime     = 2 * time.Second
	localRelayerCheckPoolTime = 100 * time.Millisecond
	localRelayerCheckTimeout  = 3 * time.Second
)

func getAssetName(version string) (string, error) {
	goarch, goos := runtime.GOARCH, runtime.GOOS
	if goos != "linux" && goos != "darwin" {
		return "", fmt.Errorf("OS not supported: %s", goos)
	}
	trimmedVersion := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("%s_%s_%s_%s.tar.gz",
		constants.RelayerRepoName,
		trimmedVersion,
		goos,
		goarch,
	), nil
}

func InstallLatest(baseDir string, authToken string) (string, error) {
	return install.InstallGithubRelease(
		constants.AvaLabsOrg,
		constants.RelayerRepoName,
		authToken,
		install.LatestRelease,
		"",
		getAssetName,
		install.TarGz,
		baseDir,
		constants.RelayerBinName,
	)
}

func InstallCustomVersion(baseDir string, version string) (string, error) {
	return install.InstallGithubRelease(
		constants.AvaLabsOrg,
		constants.RelayerRepoName,
		"",
		install.CustomRelease,
		version,
		getAssetName,
		install.TarGz,
		baseDir,
		constants.RelayerBinName,
	)
}

func Execute(
	binPath string,
	configPath string,
	logFilePath string,
	runFilePath string,
) (int, error) {
	logWriter, err := os.Create(logFilePath)
	if err != nil {
		return 0, err
	}
	args := []string{"--config-file", configPath}
	return process.Execute(binPath, args, logWriter, logWriter, runFilePath, localRelayerSetupTime)
}

func WaitForInitialization(
	configPath string,
	logPath string,
	checkInterval time.Duration,
	checkTimeout time.Duration,
) error {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	config, err := relayer.UnserializeRelayerConfig(configBytes)
	if err != nil {
		return err
	}
	sourceBlockchains := []string{}
	for _, source := range config.SourceBlockchains {
		sourceBlockchains = append(sourceBlockchains, source.BlockchainID)
	}
	if checkInterval == 0 {
		checkInterval = 100 * time.Millisecond
	}
	if checkTimeout == 0 {
		checkTimeout = 10 * time.Second
	}
	t0 := time.Now()
	for {
		bs, err := os.ReadFile(logPath)
		if err != nil {
			return err
		}
		sourcesInitialized := 0
		for _, l := range strings.Split(string(bs), "\n") {
			for _, sourceBlockchain := range sourceBlockchains {
				if strings.Contains(l, "Listener initialized") && strings.Contains(l, sourceBlockchain) {
					sourcesInitialized++
				}
			}
		}
		if sourcesInitialized == len(sourceBlockchains) {
			break
		}
		elapsed := time.Since(t0)
		if elapsed > checkTimeout {
			return fmt.Errorf("timeout waiting for relayer initialization")
		}
		time.Sleep(checkInterval)
	}
	return nil
}

func IsRunning(pid int, runFilePath string) (bool, int, *os.Process, error) {
	return process.IsRunning(pid, runFilePath)
}

func Cleanup(pid int, runFilePath string, storageDir string) error {
	return process.Cleanup(pid, runFilePath, storageDir, localRelayerCheckPoolTime, localRelayerCheckTimeout)
}
