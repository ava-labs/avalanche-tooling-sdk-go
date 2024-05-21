// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

type VMType string

const (
	SubnetEvm = "Subnet-EVM"
	CustomVM  = "Custom"
)

func VMTypeFromString(s string) VMType {
	switch s {
	case SubnetEvm:
		return SubnetEvm
	default:
		return CustomVM
	}
}

func (v VMType) RepoName() string {
	switch v {
	case SubnetEvm:
		return SubnetEVMRepoName
	default:
		return "unknown"
	}
}
