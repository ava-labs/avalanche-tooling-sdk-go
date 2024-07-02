// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import "github.com/ava-labs/avalanche-tooling-sdk-go/constants"

type VMType string

const (
	SubnetEvm = "Subnet-EVM"
)

func (v VMType) RepoName() string {
	switch v {
	case SubnetEvm:
		return constants.SubnetEVMRepoName
	default:
		return "unknown"
	}
}
