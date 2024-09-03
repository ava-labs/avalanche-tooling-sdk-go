// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

func (h *Node) PushComposeFile(localFile string, remoteFile string, merge bool) error {
	if !utils.FileExists(localFile) {
		return fmt.Errorf("file %s does not exist to be uploaded to node: %s", localFile, h.NodeID)
	}
	if err := h.MkdirAll(filepath.Dir(remoteFile), constants.SSHFileOpsTimeout); err != nil {
		return err
	}
	fileExists, err := h.FileExists(remoteFile)
	if err != nil {
		return err
	}
	h.Logger.Infof("Pushing compose file %s to %s:%s", localFile, h.NodeID, remoteFile)
	if fileExists && merge {
		// upload new and merge files
		h.Logger.Infof("Merging compose files")
		tmpFile, err := h.CreateTempFile()
		if err != nil {
			return err
		}
		defer func() {
			if err := h.Remove(tmpFile, false); err != nil {
				h.Logger.Errorf("Error removing temporary file %s:%s %s", h.NodeID, tmpFile, err)
			}
		}()
		if err := h.Upload(localFile, tmpFile, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
		if err := h.MergeComposeFiles(remoteFile, tmpFile); err != nil {
			return err
		}
	} else {
		h.Logger.Infof("Uploading compose file for node; %s", h.NodeID)
		if err := h.Upload(localFile, remoteFile, constants.SSHFileOpsTimeout); err != nil {
			return err
		}
	}
	return nil
}

// mergeComposeFiles merges two docker-compose files on a remote node.
func (h *Node) MergeComposeFiles(currentComposeFile string, newComposeFile string) error {
	fileExists, err := h.FileExists(currentComposeFile)
	if err != nil {
		return err
	}
	if !fileExists {
		return fmt.Errorf("file %s does not exist", currentComposeFile)
	}

	fileExists, err = h.FileExists(newComposeFile)
	if err != nil {
		return err
	}
	if !fileExists {
		return fmt.Errorf("file %s does not exist", newComposeFile)
	}

	output, err := h.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s -f %s config", currentComposeFile, newComposeFile)
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
	h.Logger.Infof("Merged compose files as %s", output)
	if err := h.PushComposeFile(tmpFile.Name(), currentComposeFile, false); err != nil {
		return err
	}
	return nil
}

func (h *Node) StartDockerCompose(timeout time.Duration) error {
	// we provide systemd service unit for docker compose if the node has systemd
	if h.HasSystemDAvailable() {
		if output, err := h.Command(nil, timeout, "sudo systemctl start avalanche-cli-docker"); err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	} else {
		composeFile := utils.GetRemoteComposeFile()
		output, err := h.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s up -d", composeFile)
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}
	return nil
}

func (h *Node) StopDockerCompose(timeout time.Duration) error {
	if h.HasSystemDAvailable() {
		if output, err := h.Command(nil, timeout, "sudo systemctl stop avalanche-cli-docker"); err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	} else {
		composeFile := utils.GetRemoteComposeFile()
		output, err := h.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s down", composeFile)
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}
	return nil
}

func (h *Node) RestartDockerCompose(timeout time.Duration) error {
	if h.HasSystemDAvailable() {
		if output, err := h.Command(nil, timeout, "sudo systemctl restart avalanche-cli-docker"); err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	} else {
		composeFile := utils.GetRemoteComposeFile()
		output, err := h.Commandf(nil, constants.SSHScriptTimeout, "docker compose -f %s restart", composeFile)
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}
	return nil
}

func (h *Node) StartDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if err := h.InitDockerComposeService(composeFile, service, timeout); err != nil {
		return err
	}
	if output, err := h.Commandf(nil, timeout, "docker compose -f %s start %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (h *Node) StopDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if output, err := h.Commandf(nil, timeout, "docker compose -f %s stop %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (h *Node) RestartDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if output, err := h.Commandf(nil, timeout, "docker compose -f %s restart %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (h *Node) InitDockerComposeService(composeFile string, service string, timeout time.Duration) error {
	if output, err := h.Commandf(nil, timeout, "docker compose -f %s create %s", composeFile, service); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// ComposeOverSSH sets up a docker-compose file on a remote node over SSH.
func (h *Node) ComposeOverSSH(
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
	h.Logger.Infof("pushComposeFile [%s]%s", h.NodeID, composeDesc)
	if err := h.PushComposeFile(tmpFile.Name(), remoteComposeFile, true); err != nil {
		return err
	}
	h.Logger.Infof("ValidateComposeFile [%s]%s", h.NodeID, composeDesc)
	if err := h.ValidateComposeFile(remoteComposeFile, timeout); err != nil {
		h.Logger.Errorf("ComposeOverSSH[%s]%s failed to validate: %v", h.NodeID, composeDesc, err)
		return err
	}
	h.Logger.Infof("StartDockerCompose [%s]%s", h.NodeID, composeDesc)
	if err := h.StartDockerCompose(timeout); err != nil {
		return err
	}
	executionTime := time.Since(startTime)
	h.Logger.Infof("ComposeOverSSH[%s]%s took %s with err: %v", h.NodeID, composeDesc, executionTime, err)
	return nil
}

// ListRemoteComposeServices lists the services in a remote docker-compose file.
func (h *Node) ListRemoteComposeServices(composeFile string, timeout time.Duration) ([]string, error) {
	output, err := h.Commandf(nil, timeout, "docker compose -f %s config --services", composeFile)
	if err != nil {
		return nil, err
	}
	return utils.CleanupStrings(strings.Split(string(output), "\n")), nil
}

// GetRemoteComposeContent gets the content of a remote docker-compose file.
func (h *Node) GetRemoteComposeContent(composeFile string, timeout time.Duration) (string, error) {
	tmpFile, err := os.CreateTemp("", "avalancecli-docker-compose-*.yml")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	if err := h.Download(composeFile, tmpFile.Name(), timeout); err != nil {
		return "", err
	}
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseRemoteComposeContent extracts a value from a remote docker-compose file.
func (h *Node) ParseRemoteComposeContent(composeFile string, pattern string, timeout time.Duration) (string, error) {
	content, err := h.GetRemoteComposeContent(composeFile, timeout)
	if err != nil {
		return "", err
	}
	return utils.ExtractPlaceholderValue(pattern, content)
}

// HasRemoteComposeService checks if a serviceis present in a remote docker-compose file.
func (h *Node) HasRemoteComposeService(composeFile string, service string, timeout time.Duration) (bool, error) {
	services, err := h.ListRemoteComposeServices(composeFile, timeout)
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

func (h *Node) ListDockerComposeImages(composeFile string, timeout time.Duration) (map[string]string, error) {
	output, err := h.Commandf(nil, timeout, "docker compose -f %s images --format json", composeFile)
	if err != nil {
		return nil, err
	}
	type dockerImages struct {
		ID         string `json:"ID"`
		Name       string `json:"ContainerName"`
		Repository string `json:"Repository"`
		Tag        string `json:"Tag"`
		Size       uint   `json:"Size"`
	}
	var images []dockerImages
	if err := json.Unmarshal(output, &images); err != nil {
		return nil, err
	}
	imageMap := make(map[string]string)
	for _, image := range images {
		imageMap[image.Repository] = image.Tag
	}
	return imageMap, nil
}

func (h *Node) GetDockerImageVersion(image string, timeout time.Duration) (string, error) {
	imageMap, err := h.ListDockerComposeImages(utils.GetRemoteComposeFile(), timeout)
	if err != nil {
		return "", err
	}
	return imageMap[image], nil
}
