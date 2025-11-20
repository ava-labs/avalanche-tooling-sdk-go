// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package constants

import "time"

const (
	APIRequestTimeout            = 10 * time.Second
	APIRequestLargeTimeout       = 30 * time.Second
	PrimaryWalletCreationTimeout = 2 * time.Minute
	UserOnlyWriteReadExecPerms   = 0o700
	WriteReadUserOnlyPerms       = 0o600
)

// Chain type constants
const (
	ChainTypePChain = "P-Chain"
)

// ChainAlias represents a blockchain alias identifier
type ChainAlias string

const (
	// PChainAlias is the alias for P Chain
	PChainAlias ChainAlias = "P"
	// XChainAlias is the alias for X Chain
	XChainAlias ChainAlias = "X"
	// CChainAlias is the alias for C Chain
	CChainAlias ChainAlias = "C"
	// UndefinedAlias is used for undefined chains
	UndefinedAlias ChainAlias = ""
)
