// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// PullDockerImage pulls a docker image on a remote node.
func (h *Node) PullDockerImage(image string) error {
	h.Logger.Infof("Pulling docker image %s on %s", image, h.NodeID)
	_, err := h.Commandf(nil, constants.SSHLongRunningScriptTimeout, "docker pull %s", image)
	return err
}

// DockerLocalImageExists checks if a docker image exists on a remote node.
func (h *Node) DockerLocalImageExists(image string) (bool, error) {
	output, err := h.Command(nil, constants.SSHLongRunningScriptTimeout, "docker images --format '{{.Repository}}:{{.Tag}}'")
	if err != nil {
		return false, err
	}
	for _, localImage := range parseDockerImageListOutput(output) {
		if localImage == image {
			return true, nil
		}
	}
	return false, nil
}

// parseDockerImageListOutput parses the output of a docker images command.
func parseDockerImageListOutput(output []byte) []string {
	return strings.Split(string(output), "\n")
}

// BuildDockerImage builds a docker image on a remote node.
func (h *Node) BuildDockerImage(image string, path string, dockerfile string) error {
	_, err := h.Commandf(nil, constants.SSHLongRunningScriptTimeout, "cd %s && docker build -q --build-arg GO_VERSION=%s -t %s -f %s .", path, constants.BuildEnvGolangVersion, image, dockerfile)
	return err
}

// BuildDockerImageFromGitRepo builds a docker image from a git repo on a remote node.
func (h *Node) BuildDockerImageFromGitRepo(image string, gitRepo string, commit string) error {
	if commit == "" {
		commit = "HEAD"
	}
	tmpDir, err := h.CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		if err := h.Remove(tmpDir, true); err != nil {
			h.Logger.Errorf("Error removing temporary directory %s: %s", tmpDir, err)
		}
	}()
	// clone the repo and checkout commit
	if _, err := h.Commandf(nil, constants.SSHLongRunningScriptTimeout, "git clone %s %s && cd %s && git checkout %s ", gitRepo, tmpDir, tmpDir, commit); err != nil {
		return err
	}
	// build the image
	if err := h.BuildDockerImage(image, tmpDir, "Dockerfile"); err != nil {
		return err
	}
	h.Logger.Infof("Docker image %s built from %s using %s commit/branch/tag", image, gitRepo, commit)
	return nil
}

// PrepareDockerImageWithRepo prepares a docker image on a remote node.
func (h *Node) PrepareDockerImageWithRepo(image string, gitRepo string, commit string) error {
	localImageExists, _ := h.DockerLocalImageExists(image)
	if localImageExists {
		h.Logger.Infof("Docker image %s is FOUND on %s", image, h.NodeID)
		return nil
	} else {
		h.Logger.Infof("Docker image %s not found on %s, pulling it", image, h.NodeID)
		if err := h.PullDockerImage(image); err != nil {
			h.Logger.Infof("Docker image %s not found on %s, building it from %s using %s commit/branch/tag", image, h.NodeID, gitRepo, commit)
			if err := h.BuildDockerImageFromGitRepo(image, gitRepo, commit); err != nil {
				return err
			}
			return nil
		}
	}
	h.Logger.Infof("Docker image %s is READY on %s", image, h.NodeID)
	return nil
}
