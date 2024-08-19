// Copyright (C) 2022, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.
package relayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ava-labs/avalanche-cli/pkg/application"
	"github.com/ava-labs/avalanche-cli/pkg/binutils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

const (
	localRelayerCheckPoolTime = 100 * time.Millisecond
	localRelayerCheckTimeout  = 3 * time.Second
)

type relayerRunFile struct {
	Pid int `json:"pid"`
}

func GetGithubReleaseURL(version string) (string, error) {
	goarch, goos := runtime.GOARCH, runtime.GOOS
	if goos != "linux" && goos != "darwin" {
		return "", fmt.Errorf("OS not supported: %s", goos)
	}
	trimmedVersion := strings.TrimPrefix(version, "v")
	return fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/awm-relayer_%s_%s_%s.tar.gz",
		constants.AvaLabsOrg,
		constants.RelayerRepoName,
		version,
		trimmedVersion,
		goos,
		goarch,
	), nil
}

// TODO: do something as simple as this, but more generic,
// (avago/subnet-evm/relayer/etc):
// - local install in subdir given by tool name and version
// - specify latest release, latest prerelease, or just some version
func InstallRelayer(binDir, version string) (string, error) {
	binPath := filepath.Join(binDir, constants.RelayerBinName)
	if utils.IsExecutable(binPath) {
		return binPath, nil
	}
	url, err := GetGithubReleaseURL(version)
	if err != nil {
		return "", err
	}
	bs, err := utils.HTTPGet(url, "")
	if err != nil {
		return "", err
	}
	if err := utils.InstallArchive("tar.gz", bs, binDir); err != nil {
		return "", err
	}
	return binPath, nil
}

func InstallLatestRelayer(binDir string) (string, error) {
	version, err := utils.GetLatestGithubReleaseVersion(constants.AvaLabsOrg, constants.RelayerRepoName, "")
	if err != nil {
		return "", err
	}
	versionBinDir := filepath.Join(binDir, version)
	return InstallRelayer(versionBinDir, version)
}

func DeployRelayer(
	binDir string,
	configPath string,
	logFilePath string,
	runFilePath string,
	storageDir string,
) error {
	if err := RelayerCleanup(runFilePath, storageDir); err != nil {
		return err
	}
	downloader := application.NewDownloader()
	version, err := downloader.GetLatestReleaseVersion(binutils.GetGithubLatestReleaseURL(constants.AvaLabsOrg, constants.AWMRelayerRepoName))
	if err != nil {
		return err
	}
	versionBinDir := filepath.Join(binDir, version)
	binPath, err := installRelayer(versionBinDir, version)
	if err != nil {
		return err
	}
	pid, err := executeRelayer(binPath, configPath, logFilePath)
	if err != nil {
		return err
	}
	return saveRelayerRunFile(runFilePath, pid)
}

func RelayerIsUp(runFilePath string) (bool, int, *os.Process, error) {
	if !utils.FileExists(runFilePath) {
		return false, 0, nil, nil
	}
	bs, err := os.ReadFile(runFilePath)
	if err != nil {
		return false, 0, nil, err
	}
	rf := relayerRunFile{}
	if err := json.Unmarshal(bs, &rf); err != nil {
		return false, 0, nil, err
	}
	proc, err := os.FindProcess(rf.Pid)
	if err != nil {
		// after a reboot without network cleanup, it is expected that the file pid will exist but the process not
		err := removeRelayerRunFile(runFilePath)
		return false, 0, nil, err
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		// after a reboot without network cleanup, it is expected that the file pid will exist but the process not
		// sometimes FindProcess returns without error, but Signal 0 will surely fail if the process doesn't exist
		err := removeRelayerRunFile(runFilePath)
		return false, 0, nil, err
	}
	return true, rf.Pid, proc, nil
}

func RelayerCleanup(runFilePath string, storageDir string) error {
	if err := os.RemoveAll(storageDir); err != nil {
		return err
	}
	relayerIsUp, pid, proc, err := RelayerIsUp(runFilePath)
	if err != nil {
		return err
	}
	if relayerIsUp {
		waitedCh := make(chan struct{})
		go func() {
			for {
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					if errors.Is(err, os.ErrProcessDone) {
						close(waitedCh)
						return
					} else {
						fmt.Println("failure checking to process pid %d aliveness due to: %s", proc.Pid, err)
					}
				}
				time.Sleep(localRelayerCheckPoolTime)
			}
		}()
		if err := proc.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed sending interrupt signal to relayer process with pid %d: %w", pid, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), localRelayerCheckTimeout)
		defer cancel()
		select {
		case <-ctx.Done():
			if err := proc.Signal(os.Kill); err != nil {
				return fmt.Errorf("failed killing relayer process with pid %d: %w", pid, err)
			}
		case <-waitedCh:
		}
		return removeRelayerRunFile(runFilePath)
	}
	return nil
}

func removeRelayerRunFile(runFilePath string) error {
	err := os.Remove(runFilePath)
	if err != nil {
		err = fmt.Errorf("failed removing relayer run file %s: %w", runFilePath, err)
	}
	return err
}

func saveRelayerRunFile(runFilePath string, pid int) error {
	rf := relayerRunFile{
		Pid: pid,
	}
	bs, err := json.Marshal(&rf)
	if err != nil {
		return err
	}
	if err := os.WriteFile(runFilePath, bs, constants.WriteReadReadPerms); err != nil {
		return fmt.Errorf("could not write awm relater run file to %s: %w", runFilePath, err)
	}
	return nil
}

func executeRelayer(binPath string, configPath string, logFile string) (int, error) {
	logWriter, err := os.Create(logFile)
	if err != nil {
		return 0, err
	}

	cmd := exec.Command(binPath, "--config-file", configPath)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	return cmd.Process.Pid, nil
}
