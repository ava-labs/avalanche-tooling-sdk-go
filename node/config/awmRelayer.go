// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package services

import (
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

const DockerAWMRelayerPath = "/.awm-relayer"

func GetRemoteAMWRelayerConfig() string {
	return filepath.Join(constants.CloudNodeAWMRelayerPath, constants.AWMRelayerConfigFilename)
}

func GetDockerAWMRelayerFolder() string {
	return filepath.Join(DockerAWMRelayerPath, "storage")
}

func AWMRelayerFoldersToCreate() []string {
	return []string{
		filepath.Join(constants.CloudNodeAWMRelayerPath, "storage"),
	}
}
