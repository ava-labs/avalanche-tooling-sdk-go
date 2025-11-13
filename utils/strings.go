// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// Removes the leading 0x/0X part of a hexadecimal string representation
func TrimHexa(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
}

// Formats an amount of base units as a string representing the amount in the given denomination.
// (i.e. An amount of 54321 with a decimals value of 3 results in the stirng "54.321")
func FormatAmount(amount *big.Int, decimals uint8) string {
	amountFloat := new(big.Float).SetInt(amount)
	divisor := new(big.Float).SetFloat64(math.Pow10(int(decimals)))
	val := new(big.Float).Quo(amountFloat, divisor)
	return fmt.Sprintf("%.*f", decimals, val)
}
