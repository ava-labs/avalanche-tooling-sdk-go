// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package xchain

import (
	"testing"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/vms/components/avax"
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
			name: "CreateAssetTx",
			createTx: func() txs.UnsignedTx {
				return &txs.CreateAssetTx{
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
			name: "OperationTx",
			createTx: func() txs.UnsignedTx {
				return &txs.OperationTx{
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

func TestIsXChainTx(t *testing.T) {
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
			result := IsXChainTx(tt.input)
			require.Equal(t, tt.expected, result)
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
			name: "Cascade CreateAssetTx",
			createTx: func() txs.UnsignedTx {
				return &txs.CreateAssetTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: constants.CascadeID,
						},
					},
				}
			},
			expectedHRP: constants.CascadeHRP,
			expectError: false,
		},
		{
			name: "Denali ImportTx",
			createTx: func() txs.UnsignedTx {
				return &txs.ImportTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: constants.DenaliID,
						},
					},
				}
			},
			expectedHRP: constants.DenaliHRP,
			expectError: false,
		},
		{
			name: "Everest ExportTx",
			createTx: func() txs.UnsignedTx {
				return &txs.ExportTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: constants.EverestID,
						},
					},
				}
			},
			expectedHRP: constants.EverestHRP,
			expectError: false,
		},
		{
			name: "Local OperationTx",
			createTx: func() txs.UnsignedTx {
				return &txs.OperationTx{
					BaseTx: txs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: constants.LocalID,
						},
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
						NetworkID: 88888, // Unknown network
					},
				}
			},
			expectedHRP: constants.FallbackHRP,
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
