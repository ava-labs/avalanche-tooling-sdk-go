// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package docker

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

type dockerComposeInputs struct {
	WithMonitoring     bool
	WithAvalanchego    bool
	AvalanchegoVersion string
	E2E                bool
	E2EIP              string
	E2ESuffix          string
}

//go:embed templates/*.docker-compose.yml
var composeTemplate embed.FS

func renderComposeFile(composePath string, composeDesc string, templateVars dockerComposeInputs) ([]byte, error) {
	compose, err := composeTemplate.ReadFile(composePath)
	if err != nil {
		return nil, err
	}
	var composeBytes bytes.Buffer
	t, err := template.New(composeDesc).Parse(string(compose))
	if err != nil {
		return nil, err
	}
	if err := t.Execute(&composeBytes, templateVars); err != nil {
		return nil, err
	}
	return composeBytes.Bytes(), nil
}

func (dh *DockerHost) PushComposeFile(localFile string, remoteFile string, merge bool) error {
	if !utils.FileExists(localFile) {
		return fmt.Errorf("file %s does not exist to be uploaded to host: %s", localFile, dh.Host.NodeID)
	}
	if err := dh.Host.MkdirAll(filepath.Dir(remoteFile), constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	fileExists, err := dh.Host.FileExists(remoteFile)
	if err != nil {
		return err
	}
	dh.Logger.Infof("Pushing compose file %s to %s:%s", localFile, dh.Host.NodeID, remoteFile)
	if fileExists && merge {
		// upload new and merge files
		dh.Logger.Infof("Merging compose files")
		tmpFile, err := dh.Host.CreateTempFile()
		if err != nil {
			return err
		}
		defer func() {
			if err := dh.Host.Remove(tmpFile, false); err != nil {
				dh.Logger.Errorf("Error removing temporary file %s:%s %s", dh.Host.NodeID, tmpFile, err)
			}
		}()
		if err := dh.Host.Upload(localFile, tmpFile, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
		if err := dh.MergeComposeFiles(remoteFile, tmpFile); err != nil {
			return err
		}
	} else {
		dh.Logger.Infof("Uploading compose file for host; %s", dh.Host.NodeID)
		if err := dh.Host.Upload(localFile, remoteFile, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	return nil
}

// mergeComposeFiles merges two docker-compose files on a remote host.
func (dh *DockerHost) MergeComposeFiles(currentComposeFile string, newComposeFile string) error {
	fileExists, err := dh.Host.FileExists(currentComposeFile)
	if err != nil {
		return err
	}
	if !fileExists {
		return fmt.Errorf("file %s does not exist", currentComposeFile)
	}

	fileExists, err = dh.Host.FileExists(newComposeFile)
	if err != nil {
		return err
	}
	if !fileExists {
		return fmt.Errorf("file %s does not exist", newComposeFile)
	}

	output, err := dh.Host.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s -f %s config", currentComposeFile, newComposeFile)
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	tmpFile, err := os.CreateTemp("", "avalancecli-docker-compose-*.yml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(output); err != nil {
		return err
	}
	dh.Logger.Infof("Merged compose files as %s", output)
	if err := dh.PushComposeFile(tmpFile.Name(), currentComposeFile, false); err != nil {
		return err
	}
	return nil
}

func (dh *DockerHost) StartDockerCompose(timeout time.Duration) error {
	// we provide systemd service unit for docker compose if the host has systemd
	if dh.Host.IsSystemD() {
		if output, err := dh.Host.Command(nil, timeout, "sudo systemctl start avalanche-cli-docker"); err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	} else {
		composeFile := utils.GetRemoteComposeFile()
		output, err := dh.Host.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s up -d", composeFile)
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}
	return nil
}

func (dh *DockerHost) StopDockerCompose(timeout time.Duration) error {
	if dh.Host.IsSystemD() {
		if output, err := dh.Host.Command(nil, timeout, "sudo systemctl stop avalanche-cli-docker"); err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	} else {
		composeFile := utils.GetRemoteComposeFile()
		output, err := dh.Host.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s down", composeFile)
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}
	return nil
}

func (dh *DockerHost) RestartDockerCompose(timeout time.Duration) error {
	if dh.Host.IsSystemD() {
		if output, err := dh.Host.Command(nil, timeout, "sudo systemctl restart avalanche-cli-docker"); err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	} else {
		composeFile := utils.GetRemoteComposeFile()
		output, err := dh.Host.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s restart", composeFile)
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}
	return nil
}

func (dh *DockerHost) StartDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if err := dh.InitDockerComposeService(composeFile, service, timeout); err != nil {
		return err
	}
	if output, err := dh.Host.Commandf(nil, timeout, "docker compose -f %s start %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (dh *DockerHost) StopDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if output, err := dh.Host.Commandf(nil, timeout, "docker compose -f %s stop %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (dh *DockerHost) RestartDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if output, err := dh.Host.Commandf(nil, timeout, "docker compose -f %s restart %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (dh *DockerHost) InitDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if output, err := dh.Host.Commandf(nil, timeout, "docker compose -f %s create %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// ComposeOverSSH sets up a docker-compose file on a remote host over SSH.
func (dh *DockerHost) ComposeOverSSH(
	composeDesc string,
	timeout time.Duration,
	composePath string,
	composeVars dockerComposeInputs,
) error {
	remoteComposeFile := utils.GetRemoteComposeFile()
	startTime := time.Now()
	tmpFile, err := os.CreateTemp("", "avalanchecli-docker-compose-*.yml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	composeData, err := renderComposeFile(composePath, composeDesc, composeVars)
	if err != nil {
		return err
	}

	if _, err := tmpFile.Write(composeData); err != nil {
		return err
	}
	dh.Logger.Infof("pushComposeFile [%s]%s", dh.Host.NodeID, composeDesc)
	if err := dh.PushComposeFile(tmpFile.Name(), remoteComposeFile, true); err != nil {
		return err
	}
	dh.Logger.Infof("ValidateComposeFile [%s]%s", dh.Host.NodeID, composeDesc)
	if err := dh.ValidateComposeFile(remoteComposeFile, timeout); err != nil {
		dh.Logger.Errorf("ComposeOverSSH[%s]%s failed to validate: %v", dh.Host.NodeID, composeDesc, err)
		return err
	}
	dh.Logger.Infof("StartDockerCompose [%s]%s", dh.Host.NodeID, composeDesc)
	if err := dh.StartDockerCompose(timeout); err != nil {
		return err
	}
	executionTime := time.Since(startTime)
	dh.Logger.Infof("ComposeOverSSH[%s]%s took %s with err: %v", dh.Host.NodeID, composeDesc, executionTime, err)
	return nil
}

// ListRemoteComposeServices lists the services in a remote docker-compose file.
func (dh *DockerHost) ListRemoteComposeServices(composeFile string, timeout time.Duration) ([]string, error) {
	output, err := dh.Host.Commandf(nil, timeout, "docker compose -f %s config --services", composeFile)
	if err != nil {
		return nil, err
	}
	return utils.CleanupStrings(strings.Split(string(output), "\n")), nil
}

// GetRemoteComposeContent gets the content of a remote docker-compose file.
func (dh *DockerHost) GetRemoteComposeContent(composeFile string, timeout time.Duration) (string, error) {
	tmpFile, err := os.CreateTemp("", "avalancecli-docker-compose-*.yml")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	if err := dh.Host.Download(composeFile, tmpFile.Name(), timeout); err != nil {
		return "", err
	}
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseRemoteComposeContent extracts a value from a remote docker-compose file.
func (dh *DockerHost) ParseRemoteComposeContent(composeFile string, pattern string, timeout time.Duration) (string, error) {
	content, err := dh.GetRemoteComposeContent(composeFile, timeout)
	if err != nil {
		return "", err
	}
	return utils.ExtractPlaceholderValue(pattern, content)
}

// HasRemoteComposeService checks if a serviceis present in a remote docker-compose file.
func (dh *DockerHost) HasRemoteComposeService(composeFile string, service string, timeout time.Duration) (bool, error) {
	services, err := dh.ListRemoteComposeServices(composeFile, timeout)
	if err != nil {
		return false, err
	}
	found := false
	for _, s := range services {
		if s == service {
			found = true
			break
		}
	}
	return found, nil
}
