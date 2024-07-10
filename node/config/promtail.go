// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package services

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

func PromtailFoldersToCreate() []string {
	return []string{
		utils.GetRemoteComposeServicePath(constants.ServicePromtail),
		"/home/ubuntu/.avalanchego/logs",
	}
}