// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"net/url"
)

// IsEndpointLocalhost checks if the endpoint is localhost or empty
func IsEndpointLocalhost(endpoint string) bool {
	if endpoint == "" {
		return true // empty means it will use local endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}

	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
