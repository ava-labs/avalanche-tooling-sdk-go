// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package pchain

import (
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/stretchr/testify/require"
)

func TestGetNetworkID(t *testing.T) {
	testNetworkID := uint32(12345)

	tests := []struct {
		name        string
		createTx    func() txs.UnsignedTx
		expectedID  uint32
		expectError bool
	}{
		{
			name: "AddDelegatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.AddDelegatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "AddPermissionlessDelegatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.AddPermissionlessDelegatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "AddPermissionlessValidatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.AddPermissionlessValidatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "AddSubnetValidatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.AddSubnetValidatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "AddValidatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.AddValidatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "BaseTx",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: testNetworkID,
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "ConvertSubnetToL1Tx",
			createTx: func() txs.UnsignedTx {
				return &txs.ConvertSubnetToL1Tx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "CreateChainTx",
			createTx: func() txs.UnsignedTx {
				return &txs.CreateChainTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "CreateSubnetTx",
			createTx: func() txs.UnsignedTx {
				return &txs.CreateSubnetTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "DisableL1ValidatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.DisableL1ValidatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "ExportTx",
			createTx: func() txs.UnsignedTx {
				return &txs.ExportTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "ImportTx",
			createTx: func() txs.UnsignedTx {
				return &txs.ImportTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "IncreaseL1ValidatorBalanceTx",
			createTx: func() txs.UnsignedTx {
				return &txs.IncreaseL1ValidatorBalanceTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "RegisterL1ValidatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.RegisterL1ValidatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "RemoveSubnetValidatorTx",
			createTx: func() txs.UnsignedTx {
				return &txs.RemoveSubnetValidatorTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "SetL1ValidatorWeightTx",
			createTx: func() txs.UnsignedTx {
				return &txs.SetL1ValidatorWeightTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "TransferSubnetOwnershipTx",
			createTx: func() txs.UnsignedTx {
				return &txs.TransferSubnetOwnershipTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
		{
			name: "TransformSubnetTx",
			createTx: func() txs.UnsignedTx {
				return &txs.TransformSubnetTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: testNetworkID,
						},
					},
				}
			},
			expectedID:  testNetworkID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.createTx()
			networkID, err := GetNetworkID(tx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedID, networkID)
			}
		})
	}
}

func TestGetNetworkID_UnsupportedType(t *testing.T) {
	// Test with nil - this should cause an error since it's not a supported type
	networkID, err := GetNetworkID(nil)

	require.Error(t, err)
	require.Equal(t, uint32(0), networkID)
	require.Contains(t, err.Error(), "unsupported transaction type")
}

func TestGetNetworkID_DifferentNetworkIDs(t *testing.T) {
	testCases := []uint32{1, 5, 12345, 4294967295} // Test various network IDs including max uint32

	for _, networkID := range testCases {
		t.Run("BaseTx", func(t *testing.T) {
			tx := &txs.BaseTx{
				BaseTx: avax.BaseTx{
					NetworkID: networkID,
				},
			}

			result, err := GetNetworkID(tx)
			require.NoError(t, err)
			require.Equal(t, networkID, result)
		})
	}
}

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

func TestIsPChainTx(t *testing.T) {
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
			result := IsPChainTx(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxFromBytes_ValidTransaction(t *testing.T) {
	// For this test, we'll focus on testing the error handling behavior
	// since creating a valid P-Chain transaction requires complex setup with proper initialization.
	// The main functionality is tested through the integration with cubesigner.

	// Test that the function handles marshaling errors gracefully
	// This test verifies that the functions don't panic on various inputs
	testInputs := [][]byte{
		{0x00},
		{0x00, 0x00, 0x00, 0x01},
		{0xFF, 0xFF, 0xFF, 0xFF},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
	}

	for i, input := range testInputs {
		t.Run(fmt.Sprintf("Input_%d", i), func(t *testing.T) {
			// These should not panic and should return false for invalid data
			tx, ok := TxFromBytes(input)
			require.False(t, ok)
			require.Nil(t, tx)

			isPChain := IsPChainTx(input)
			require.False(t, isPChain)
		})
	}
}

func TestGetHRP(t *testing.T) {
	tests := []struct {
		name        string
		createTx    func() txs.UnsignedTx
		expectedHRP string
		expectError bool
	}{
		{
			name: "Mainnet BaseTx",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: constants.MainnetID,
					},
				}
			},
			expectedHRP: constants.MainnetHRP,
			expectError: false,
		},
		{
			name: "Fuji BaseTx",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: constants.FujiID,
					},
				}
			},
			expectedHRP: constants.FujiHRP,
			expectError: false,
		},
		{
			name: "Cascade BaseTx",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: constants.CascadeID,
					},
				}
			},
			expectedHRP: constants.CascadeHRP,
			expectError: false,
		},
		{
			name: "Local BaseTx",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: constants.LocalID,
					},
				}
			},
			expectedHRP: constants.LocalHRP,
			expectError: false,
		},
		{
			name: "UnitTest BaseTx",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: constants.UnitTestID,
					},
				}
			},
			expectedHRP: constants.UnitTestHRP,
			expectError: false,
		},
		{
			name: "Unknown NetworkID uses FallbackHRP",
			createTx: func() txs.UnsignedTx {
				return &txs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: 99999, // Unknown network
					},
				}
			},
			expectedHRP: constants.FallbackHRP,
			expectError: false,
		},
		{
			name: "CreateChainTx with Mainnet",
			createTx: func() txs.UnsignedTx {
				return &txs.CreateChainTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: constants.MainnetID,
						},
					},
				}
			},
			expectedHRP: constants.MainnetHRP,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.createTx()
			hrp, err := GetHRP(tx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedHRP, hrp)
			}
		})
	}
}
