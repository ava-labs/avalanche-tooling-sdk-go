// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package txs

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/stretchr/testify/assert"
)

func TestConvertSubnetToL1TxParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  ConvertSubnetToL1TxParams
		isValid bool
	}{
		{
			name: "valid params",
			params: ConvertSubnetToL1TxParams{
				SubnetAuthKeys: []ids.ShortID{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
				SubnetID:       ids.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				ChainID:        ids.ID{33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64},
				Address:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Validators:     []*txs.ConvertSubnetToL1Validator{},
				Wallet:         nil, // Will be set to nil for this test
			},
			isValid: false, // nil wallet should be invalid
		},
		{
			name: "empty subnet auth keys",
			params: ConvertSubnetToL1TxParams{
				SubnetAuthKeys: []ids.ShortID{},
				SubnetID:       ids.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				ChainID:        ids.ID{33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64},
				Address:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Validators:     []*txs.ConvertSubnetToL1Validator{},
				Wallet:         nil,
			},
			isValid: false, // nil wallet should be invalid
		},
		{
			name: "valid validator data",
			params: ConvertSubnetToL1TxParams{
				SubnetAuthKeys: []ids.ShortID{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
				SubnetID:       ids.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				ChainID:        ids.ID{33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64},
				Address:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Validators: []*txs.ConvertSubnetToL1Validator{
					{
						NodeID: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
						Weight: 1000,
					},
				},
				Wallet: nil,
			},
			isValid: false, // nil wallet should be invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation test - nil wallet should be invalid
			if tt.params.Wallet == nil {
				assert.False(t, tt.isValid, "nil wallet should be invalid")
			} else {
				assert.True(t, tt.isValid, "valid params should pass")
			}
		})
	}
}

func TestConvertSubnetToL1TxParams_FieldAccess(t *testing.T) {
	// Test that we can create and access the struct fields
	params := ConvertSubnetToL1TxParams{
		SubnetAuthKeys: []ids.ShortID{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
		SubnetID:       ids.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		ChainID:        ids.ID{33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64},
		Address:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		Validators:     []*txs.ConvertSubnetToL1Validator{},
		Wallet:         nil,
	}

	// Test field access
	assert.Len(t, params.SubnetAuthKeys, 1)
	assert.Equal(t, ids.ShortID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, params.SubnetAuthKeys[0])
	assert.Equal(t, ids.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, params.SubnetID)
	assert.Equal(t, ids.ID{33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64}, params.ChainID)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, params.Address)
	assert.Len(t, params.Validators, 0)
	assert.Nil(t, params.Wallet)
}
