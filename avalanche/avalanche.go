// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

const (
	SubnetEVMRepoName = "subnet-evm"
)

// BaseApp provides a base type that implements the foundation of Avalanche SDK
// including Logger
type BaseApp struct {
	// The logger writer interface to write logging messages to.
	Logger LeveledLoggerInterface
}

// New creates a new base type of Avalanche SDK
// if logger is nil, default logger DefaultLeveledLogger will be used instead
func New(logger LeveledLoggerInterface) *BaseApp {
	if logger == nil {
		logger = DefaultLeveledLogger
	}
	baseApp := &BaseApp{
		Logger: logger,
	}
	return baseApp
}
