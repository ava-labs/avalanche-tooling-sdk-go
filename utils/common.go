// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

func Belongs[T comparable](input []T, elem T) bool {
	for _, e := range input {
		if e == elem {
			return true
		}
	}
	return false
}

// AppendSlices appends multiple slices into a single slice.
func AppendSlices[T any](slices ...[]T) []T {
	totalLength := 0
	for _, slice := range slices {
		totalLength += len(slice)
	}
	result := make([]T, 0, totalLength)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}
