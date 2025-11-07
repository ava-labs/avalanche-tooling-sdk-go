// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

// Option configures wallet operations
type Option func(*Options)

// Options holds configuration for wallet operations
type Options struct {
	AccountName string           // Named account from wallet
	Address     string           // Explicit address (overrides AccountName)
	WarpMessage interface{}      // Warp message for cross-chain operations (*warp.Message)
	ErrorMap    map[string]error // Maps Solidity error signatures to Go errors
}

// WithAccount specifies a named account from the wallet
func WithAccount(accountName string) Option {
	return func(opts *Options) {
		opts.AccountName = accountName
	}
}

// WithAddress specifies an explicit address
func WithAddress(address string) Option {
	return func(opts *Options) {
		opts.Address = address
	}
}

// WithWarpMessage specifies a warp message for cross-chain contract calls
func WithWarpMessage(warpMessage interface{}) Option {
	return func(opts *Options) {
		opts.WarpMessage = warpMessage
	}
}

// WithErrorMap specifies custom error mappings for Solidity errors
func WithErrorMap(errorMap map[string]error) Option {
	return func(opts *Options) {
		opts.ErrorMap = errorMap
	}
}
