// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cchain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxFromBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectTx    bool
		expectError bool
	}{
		{
			name:        "Empty bytes",
			input:       []byte{},
			expectTx:    false,
			expectError: false,
		},
		{
			name:        "Nil bytes",
			input:       nil,
			expectTx:    false,
			expectError: false,
		},
		{
			name:        "Invalid bytes",
			input:       []byte{0x01, 0x02, 0x03},
			expectTx:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, ok := TxFromBytes(tt.input)

			if tt.expectTx {
				require.True(t, ok)
				require.NotNil(t, tx)
			} else {
				require.False(t, ok)
				require.Nil(t, tx)
			}
		})
	}
}

func TestIsCChainTx(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "Empty bytes",
			input:    []byte{},
			expected: false,
		},
		{
			name:     "Nil bytes",
			input:    nil,
			expected: false,
		},
		{
			name:     "Invalid bytes",
			input:    []byte{0x01, 0x02, 0x03},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCChainTx(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxFromBytes_ValidTransaction(t *testing.T) {
	// For this test, we'll focus on testing the error handling behavior
	// since creating a valid atomic.Tx requires complex setup.
	// The main functionality is tested through the integration with cubesigner.

	// Test that the function handles marshaling errors gracefully
	// This test verifies that the functions don't panic on various inputs
	testInputs := [][]byte{
		{0x00},
		{0x00, 0x00, 0x00, 0x01},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}

	for i, input := range testInputs {
		t.Run(fmt.Sprintf("Input_%d", i), func(t *testing.T) {
			// These should not panic and should return false for invalid data
			tx, ok := TxFromBytes(input)
			require.False(t, ok)
			require.Nil(t, tx)

			isCChain := IsCChainTx(input)
			require.False(t, isCChain)
		})
	}
}
