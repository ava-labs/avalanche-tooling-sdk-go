// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

type SupportedCloud int

const (
	AWSCloud SupportedCloud = iota
	GCPCloud
	Docker // fake Cloud used for E2E tests
	NotSupported
)

type SupportedRole int

const (
	Validator SupportedRole = iota
	API
	AWMRelayer
	Loadtest
	Monitor
)
