// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

const (
	SubnetEVMRepoName = "subnet-evm"
)

// Client provides client interface for Avalanche operations
type Client struct {
	// The logger writer interface to write logging messages to.
	Logger LeveledLoggerInterface
}

func New() *Client {
	client := &Client{
		Logger: DefaultLeveledLogger,
	}
	return client
}