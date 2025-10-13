// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package constants

import "time"

const (
	APIRequestTimeout          = 10 * time.Second
	APIRequestLargeTimeout     = 30 * time.Second
	UserOnlyWriteReadExecPerms = 0o700
	WriteReadUserOnlyPerms     = 0o600
)
