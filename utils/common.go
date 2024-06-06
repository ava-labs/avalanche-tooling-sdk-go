// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

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

// RetryFunction retries the given function until it succeeds or the maximum number of attempts is reached.
func RetryFunction(fn func() (interface{}, error), maxAttempts int, retryInterval time.Duration) (
	interface{},
	error,
) {
	var err error
	var result interface{}
	const defaultRetryInterval = 2 * time.Second
	if retryInterval == 0 {
		retryInterval = defaultRetryInterval
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		time.Sleep(retryInterval)
	}
	return nil, fmt.Errorf("maximum retry attempts reached: %w", err)
}

// TimedFunction is a function that executes the given function `f` within a specified timeout duration.
func TimedFunction(
	f func() (interface{}, error),
	name string,
	timeout time.Duration,
) (interface{}, error) {
	var (
		ret interface{}
		err error
	)
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		ret, err = f()
		close(ch)
	}()
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%s timeout of %d seconds", name, uint(timeout.Seconds()))
	case <-ch:
	}
	return ret, err
}

// TimedFunctionWithRetry is a function that executes the given function `f` within a specified timeout duration.
func TimedFunctionWithRetry(
	f func() (interface{}, error),
	name string,
	timeout time.Duration,
	maxAttempts int,
	retryInterval time.Duration,
) (interface{}, error) {
	return RetryFunction(func() (interface{}, error) {
		return TimedFunction(f, name, timeout)
	}, maxAttempts, retryInterval)
}

// RandomString generates a random string of the specified length.
func RandomString(length int) string {
	randG := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404
	chars := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[randG.Intn(len(chars))]
	}
	return string(result)
}
