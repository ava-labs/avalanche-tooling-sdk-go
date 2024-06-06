// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"context"
	"sort"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
)

const (
	APIRequestTimeout      = 30 * time.Second
	APIRequestLargeTimeout = 2 * time.Minute
	WriteReadUserOnlyPerms = 0o600
)

func MapE[T, U any](input []T, f func(T) (U, error)) ([]U, error) {
	output := make([]U, 0, len(input))
	for _, e := range input {
		o, err := f(e)
		if err != nil {
			return nil, err
		}
		output = append(output, o)
	}
	return output, nil
}

// Unique returns a new slice containing only the unique elements from the input slice.
func Unique[T comparable](arr []T) []T {
	visited := map[T]bool{}
	unique := []T{}
	for _, e := range arr {
		if !visited[e] {
			unique = append(unique, e)
			visited[e] = true
		}
	}
	return unique
}

func Uint32Sort(arr []uint32) {
	sort.Slice(arr, func(i, j int) bool { return arr[i] < arr[j] })
}

// Context for API requests
func GetAPIContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), APIRequestTimeout)
}

// Context for API requests with large timeout
func GetAPILargeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), APIRequestLargeTimeout)
}

func P(
	networkHRP string,
	addresses []ids.ShortID,
) ([]string, error) {
	return MapE(
		addresses,
		func(addr ids.ShortID) (string, error) {
			return address.Format("P", networkHRP, addr[:])
		},
	)
}
