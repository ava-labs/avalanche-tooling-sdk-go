// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package docker

import (
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// PullDockerImage pulls a docker image on a remote host.
func (dh *DockerHost) PullDockerImage(image string) error {
	dh.Logger.Infof("Pulling docker image %s on %s", image, dh.Host.NodeID)
	_, err := dh.Host.Commandf(nil, constants.SSHLongRunningScriptTimeout, "docker pull %s", image)
	return err
}

// DockerLocalImageExists checks if a docker image exists on a remote host.
func (dh *DockerHost) DockerLocalImageExists(image string) (bool, error) {
	output, err := dh.Host.Command(nil, constants.SSHLongRunningScriptTimeout, "docker images --format '{{.Repository}}:{{.Tag}}'")
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

// BuildDockerImage builds a docker image on a remote host.
func (dh *DockerHost) BuildDockerImage(image string, path string, dockerfile string) error {
	_, err := dh.Host.Commandf(nil, constants.SSHLongRunningScriptTimeout, "cd %s && docker build -q --build-arg GO_VERSION=%s -t %s -f %s .", path, constants.BuildEnvGolangVersion, image, dockerfile)
	return err
}

// BuildDockerImageFromGitRepo builds a docker image from a git repo on a remote host.
func (dh *DockerHost) BuildDockerImageFromGitRepo(image string, gitRepo string, commit string) error {
	if commit == "" {
		commit = "HEAD"
	}
	tmpDir, err := dh.Host.CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		if err := dh.Host.Remove(tmpDir, true); err != nil {
			dh.Logger.Errorf("Error removing temporary directory %s: %s", tmpDir, err)
		}
	}()
	// clone the repo and checkout commit
	if _, err := dh.Host.Commandf(nil, constants.SSHLongRunningScriptTimeout, "git clone %s %s && cd %s && git checkout %s ", gitRepo, tmpDir, tmpDir, commit); err != nil {
		return err
	}
	// build the image
	if err := dh.BuildDockerImage(image, tmpDir, "Dockerfile"); err != nil {
		return err
	}
	dh.Logger.Infof("Docker image %s built from %s using %s commit/branch/tag", image, gitRepo, commit)
	return nil
}

// PrepareDockerImageWithRepo prepares a docker image on a remote host.
func (dh *DockerHost) PrepareDockerImageWithRepo(image string, gitRepo string, commit string) error {
	localImageExists, _ := dh.DockerLocalImageExists(image)
	if localImageExists {
		dh.Logger.Infof("Docker image %s is FOUND on %s", image, dh.Host.NodeID)
		return nil
	} else {
		dh.Logger.Infof("Docker image %s not found on %s, pulling it", image, dh.Host.NodeID)
		if err := dh.PullDockerImage(image); err != nil {
			dh.Logger.Infof("Docker image %s not found on %s, building it from %s using %s commit/branch/tag", image, dh.Host.NodeID, gitRepo, commit)
			if err := dh.BuildDockerImageFromGitRepo(image, gitRepo, commit); err != nil {
				return err
			}
			return nil
		}
	}
	dh.Logger.Infof("Docker image %s is READY on %s", image, dh.Host.NodeID)
	return nil
}
