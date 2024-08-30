// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestSplitTypes(t *testing.T) {
	type TestEsp struct {
		desc     string
		input    string
		expected []string
	}
	for _, testEsp := range []TestEsp{
		{
			desc:  "primitive type bool",
			input: "bool",
			expected: []string{
				"bool",
			},
		},
		{
			desc:  "struct that contains a bool",
			input: "(bool)",
			expected: []string{
				"(bool)",
			},
		},
		{
			desc:  "array of bools",
			input: "[bool]",
			expected: []string{
				"[bool]",
			},
		},
		{
			desc:  "array of structs that contain a bool",
			input: "[(bool)]",
			expected: []string{
				"[(bool)]",
			},
		},
		{
			desc:  "struct that contains an address and a uint256",
			input: "(address, uint256)",
			expected: []string{
				"(address, uint256)",
			},
		},
		{
			desc:  "array of structs that contain an address and a uint256",
			input: "[(address, uint256)]",
			expected: []string{
				"[(address, uint256)]",
			},
		},
		{
			desc:  "list of types address and uint256",
			input: "address, uint256",
			expected: []string{
				"address",
				"uint256",
			},
		},
		{
			desc:  "list of types with a simple struct and a uint256",
			input: "(bytes32, uint256), uint256",
			expected: []string{
				"(bytes32, uint256)",
				"uint256",
			},
		},
		{
			desc:  "list of types with a nested struct and a bool",
			input: "(bytes32, address, (address, uint256), uint256, [address], bytes), bool",
			expected: []string{
				"(bytes32, address, (address, uint256), uint256, [address], bytes)",
				"bool",
			},
		},
		{
			desc:  "list of types that includes a simple struct and an array",
			input: "bytes32, address, (address, uint256), uint256, [address], bytes",
			expected: []string{
				"bytes32",
				"address",
				"(address, uint256)",
				"uint256",
				"[address]",
				"bytes",
			},
		},
	} {
		output := splitTypes(testEsp.input)
		require.Equal(t, testEsp.expected, output, testEsp.desc)
	}
}

func TestGetABIMaps(t *testing.T) {
	type TestEsp struct {
		desc     string
		input    string
		values   interface{}
		expected []map[string]interface{}
	}
	type TeleporterFeeInfo struct {
		FeeTokenAddress common.Address
		Amount          *big.Int
	}
	type TeleporterMessageInput struct {
		DestinationBlockchainID [32]byte
		DestinationAddress      common.Address
		FeeInfo                 TeleporterFeeInfo
		RequiredGasLimit        *big.Int
		AllowedRelayerAddresses []common.Address
		Message                 []byte
	}
	type BoolInput struct {
		FieldName bool
	}
	for _, testEsp := range []TestEsp{
		{
			desc:  "primitive type bool",
			input: "bool",
			values: []interface{}{
				true,
			},
			expected: []map[string]interface{}{
				{
					"internalType": "bool",
					"name":         "",
					"type":         "bool",
				},
			},
		},
		{
			desc:  "slice of two primitives",
			input: "bool,int",
			values: []interface{}{
				true,
				5,
			},
			expected: []map[string]interface{}{
				{
					"internalType": "bool",
					"name":         "",
					"type":         "bool",
				},
				{
					"internalType": "int",
					"name":         "",
					"type":         "int",
				},
			},
		},
		{
			desc:  "struct of bool",
			input: "(bool)",
			values: []interface{}{
				BoolInput{},
			},
			expected: []map[string]interface{}{
				{
					"internalType": "tuple",
					"name":         "",
					"type":         "tuple",
					"components": []map[string]interface{}{
						{
							"internalType": "bool",
							"name":         "FieldName",
							"type":         "bool",
						},
					},
				},
			},
		},
		{
			desc:  "array of bool",
			input: "[bool]",
			values: []interface{}{
				BoolInput{},
			},
			expected: []map[string]interface{}{
				{
					"internalType": "bool[]",
					"name":         "",
					"type":         "bool[]",
				},
			},
		},
		{
			desc:  "array of struct of bool",
			input: "[(bool)]",
			values: []interface{}{
				[]BoolInput{},
			},
			expected: []map[string]interface{}{
				{
					"internalType": "struct BoolInput[]",
					"name":         "",
					"type":         "tuple[]",
					"components": []map[string]interface{}{
						{
							"internalType": "bool",
							"name":         "FieldName",
							"type":         "bool",
						},
					},
				},
			},
		},
		{
			desc:  "sendCrossChainMessage input",
			input: "(bytes32, address, (address, uint256), uint256, [address], bytes)",
			values: []interface{}{
				TeleporterMessageInput{},
			},
			expected: []map[string]interface{}{
				{
					"internalType": "tuple",
					"name":         "",
					"type":         "tuple",
					"components": []map[string]interface{}{
						{
							"internalType": "bytes32",
							"name":         "DestinationBlockchainID",
							"type":         "bytes32",
						},
						{
							"internalType": "address",
							"name":         "DestinationAddress",
							"type":         "address",
						},
						{
							"components": []map[string]interface{}{
								{
									"internalType": "address",
									"name":         "FeeTokenAddress",
									"type":         "address",
								},
								{
									"internalType": "uint256",
									"name":         "Amount",
									"type":         "uint256",
								},
							},
							"internalType": "struct TeleporterFeeInfo",
							"name":         "FeeInfo",
							"type":         "tuple",
						},
						{
							"internalType": "uint256",
							"name":         "RequiredGasLimit",
							"type":         "uint256",
						},
						{
							"internalType": "address[]",
							"name":         "AllowedRelayerAddresses",
							"type":         "address[]",
						},
						{
							"internalType": "bytes",
							"name":         "Message",
							"type":         "bytes",
						},
					},
				},
			},
		},
	} {
		output := splitTypes(testEsp.input)
		maps, err := getABIMaps(output, testEsp.values)
		require.NoError(t, err)
		require.Equal(t, testEsp.expected, maps)
	}
}
