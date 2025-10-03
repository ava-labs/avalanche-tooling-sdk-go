// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"strings"
)

// Removes the leading 0x/0X part of a hexadecimal string representation
func TrimHexa(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
}
