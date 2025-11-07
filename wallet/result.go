// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import "fmt"

// ParseResult extracts a typed value from ReadContract results
// Usage:
//
// Single return value:
//
//	result, err := w.ReadContract(contractAddr, method)
//	value, err := wallet.ParseResult[[32]byte](result)
//
// Multiple return values (specify index):
//
//	result, err := w.ReadContract(contractAddr, method)
//	addr, err := wallet.ParseResult[common.Address](result, 0)
//	amount, err := wallet.ParseResult[*big.Int](result, 1)
func ParseResult[T any](result []interface{}, index ...int) (T, error) {
	var zero T

	if len(index) == 0 {
		if len(result) != 1 {
			return zero, fmt.Errorf("expected single return value, got %d (specify index for multi-value returns)", len(result))
		}
		v, ok := result[0].(T)
		if !ok {
			return zero, fmt.Errorf("invalid type: expected %T, got %T", zero, result[0])
		}
		return v, nil
	}

	idx := index[0]
	if idx < 0 || idx >= len(result) {
		return zero, fmt.Errorf("index %d out of range for %d results", idx, len(result))
	}

	v, ok := result[idx].(T)
	if !ok {
		return zero, fmt.Errorf("invalid type for result[%d]: expected %T, got %T", idx, zero, result[idx])
	}

	return v, nil
}
