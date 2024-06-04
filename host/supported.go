// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

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

// String returns the string representation of the SupportedRole
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

// StringToCloud converts a string to a SupportedCloud
func StringToCloud(s string) SupportedCloud {
	switch s {
	case "aws":
		return AWSCloud
	case "gcp":
		return GCPCloud
	case "docker":
		return Docker
	default:
		return Unknown
	}
}

// String returns the string representation of the SupportedRole
func (r *SupportedRole) String() string {
	switch *r {
	case Validator:
		return "validator"
	case API:
		return "api"
	case AWMRelayer:
		return "awm-relayer"
	case Loadtest:
		return "loadtest"
	case Monitor:
		return "monitor"
	default:
		return "unknown"
	}
}

// StringToRole converts a string to a SupportedRole
func StringToRole(s string) SupportedRole {
	switch s {
	case "validator":
		return Validator
	case "api":
		return API
	case "awm-relayer":
		return AWMRelayer
	case "loadtest":
		return Loadtest
	case "monitor":
		return Monitor
	default:
		return Monitor
	}
}
