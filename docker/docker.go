// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package docker

import (
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/host"
)

type DockerHost struct {
	Host   host.Host
	Logger avalanche.LeveledLogger
}
