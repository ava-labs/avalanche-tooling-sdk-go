// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import "net"

func IsValidIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}
