// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package remoteconfig

import (
	"embed"

	"github.com/ava-labs/avalanche-tooling-sdk-go/internal/utils"
)

//go:embed templates/*
var templates embed.FS

// RemoteFoldersToCreateAvalanchego returns a list of folders that need to be created on the remote Avalanchego server
func RemoteFoldersToCreateAvalanchego() []string {
	return utils.AppendSlices[string](
		AvalancheFolderToCreate(),
	)
}
