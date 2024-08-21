// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"context"
	"sort"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
)

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
	return context.WithTimeout(context.Background(), constants.APIRequestTimeout)
}

// Context for API requests with large timeout
func GetAPILargeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), constants.APIRequestLargeTimeout)
}

func P(
	networkHRP string,
	addresses []ids.ShortID,
) ([]string, error) {
	return MapWithError(
		addresses,
		func(addr ids.ShortID) (string, error) {
			return address.Format("P", networkHRP, addr[:])
		},
	)
}
