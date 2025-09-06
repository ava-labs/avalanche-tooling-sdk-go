// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainType_String(t *testing.T) {
	tests := []struct {
		chainType ChainType
		expected  string
	}{
		{Undefined, "undefined"},
		{CChain, "C"},
		{XChain, "X"},
		{PChain, "P"},
		{ChainType(999), "undefined"}, // Invalid chain type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.chainType.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectTxChainType(t *testing.T) {
	tests := []struct {
		name     string
		bytes    []byte
		expected ChainType
	}{
		{
			name:     "Empty bytes",
			bytes:    []byte{},
			expected: Undefined,
		},
		{
			name:     "Nil bytes",
			bytes:    nil,
			expected: Undefined,
		},
		{
			name:     "Random bytes",
			bytes:    []byte{0x01, 0x02, 0x03, 0x04},
			expected: Undefined,
		},
		{
			name:     "Short invalid bytes",
			bytes:    []byte{0xFF},
			expected: Undefined,
		},
		{
			name:     "Longer invalid bytes",
			bytes:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			expected: Undefined,
		},
		{
			name:     "Potential codec pattern",
			bytes:    []byte{0x00, 0x00, 0x30, 0x39, 0x00, 0x00, 0x00, 0x01},
			expected: Undefined,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectTxChainType(tt.bytes)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectTxChainType_CrossChainConfusion(t *testing.T) {
	// Test patterns that might cause cross-chain confusion
	confusingPatterns := [][]byte{
		{0x00, 0x00, 0x30, 0x39}, // Common codec prefix
		{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
		{0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00},
	}

	for i, pattern := range confusingPatterns {
		t.Run(fmt.Sprintf("ConfusingPattern_%d", i), func(t *testing.T) {
			chainType := DetectTxChainType(pattern)

			// For potentially confusing patterns, should return Undefined
			// This test mainly ensures no panics occur
			t.Logf("Pattern %x detected as: %s", pattern, chainType.String())

			// Should be one of the valid chain types
			validTypes := []ChainType{Undefined, CChain, XChain, PChain}
			found := false
			for _, validType := range validTypes {
				if chainType == validType {
					found = true
					break
				}
			}
			require.True(t, found, "Should return a valid ChainType")
		})
	}
}

func TestDetectTxChainType_Performance(t *testing.T) {
	// Test with various sizes to ensure no performance issues
	testSizes := []int{0, 1, 32, 1024, 10240}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size_%d_bytes", size), func(t *testing.T) {
			// Create test data
			testData := make([]byte, size)
			for i := range testData {
				testData[i] = byte(i % 256)
			}

			// Should not panic or hang
			chainType := DetectTxChainType(testData)

			// Large invalid data should be Undefined
			require.Equal(t, Undefined, chainType, "Large invalid data should be Undefined")
		})
	}
}

func TestDetectTxChainType_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name  string
		bytes []byte
	}{
		{"Single zero byte", []byte{0x00}},
		{"Single max byte", []byte{0xFF}},
		{"All zeros", make([]byte, 100)}, // Will be all zeros by default
		{"Sequential pattern", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}()},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			chainType := DetectTxChainType(tc.bytes)

			// All edge cases should be Undefined for invalid data
			require.Equal(t, Undefined, chainType)
		})
	}
}
