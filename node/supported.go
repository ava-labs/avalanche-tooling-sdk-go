// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

type SupportedCloud int

const (
	AWSCloud SupportedCloud = iota
	GCPCloud
	Docker // fake Cloud used for E2E tests
	Unknown
)

type SupportedRole int

const (
	Validator SupportedRole = iota
	API
	AWMRelayer
	Loadtest
	Monitor
)

func (c *SupportedCloud) String() string {
	switch *c {
	case AWSCloud:
		return "aws"
	case GCPCloud:
		return "gcp"
	case Docker:
		return "docker"
	default:
		return "unknown"
	}
}
