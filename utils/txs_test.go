// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/bls/signer/localsigner"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/message"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/payload"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/chain/x/builder"
	"github.com/ava-labs/coreth/plugin/evm/atomic"
	"github.com/ava-labs/libevm/common"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"

	avmtxs "github.com/ava-labs/avalanchego/vms/avm/txs"
	platformvmtxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

func TestAutoDetectChainPChainTxs(t *testing.T) {
	require := require.New(t)

	addr := ids.ShortID{
		0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
		0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
		0x44, 0x55, 0x66, 0x77,
	}

	txID := ids.ID{
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
	}

	avaxAssetID, err := ids.FromString("FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z")
	require.NoError(err)

	baseTxData := platformvmtxs.BaseTx{
		BaseTx: avax.BaseTx{
			NetworkID:    1,
			BlockchainID: ids.GenerateTestID(),
			Ins: []*avax.TransferableInput{
				{
					UTXOID: avax.UTXOID{
						TxID:        txID,
						OutputIndex: 1,
					},
					Asset: avax.Asset{
						ID: avaxAssetID,
					},
					In: &secp256k1fx.TransferInput{
						Amt: 1000000,
						Input: secp256k1fx.Input{
							SigIndices: []uint32{0},
						},
					},
				},
			},
			Outs: []*avax.TransferableOutput{},
		},
	}

	// Generate BLS key for ConvertSubnetToL1Tx
	sk, err := localsigner.New()
	require.NoError(err)

	pop, err := signer.NewProofOfPossession(sk)
	require.NoError(err)

	validators := []*platformvmtxs.ConvertSubnetToL1Validator{
		{
			NodeID:  ids.GenerateTestNodeID().Bytes(),
			Weight:  1000,
			Balance: 100000,
			Signer:  *pop,
			RemainingBalanceOwner: message.PChainOwner{
				Threshold: 1,
				Addresses: []ids.ShortID{addr},
			},
			DeactivationOwner: message.PChainOwner{
				Threshold: 1,
				Addresses: []ids.ShortID{addr},
			},
		},
	}

	// Create warp message for RegisterL1ValidatorTx test
	subnetID := ids.GenerateTestID()
	nodeID := ids.GenerateTestNodeID()
	blsPublicKey := [48]byte{}
	expiry := uint64(1234567890)
	weight := uint64(1000)

	balanceOwners := message.PChainOwner{
		Threshold: 1,
		Addresses: []ids.ShortID{addr},
	}
	disableOwners := message.PChainOwner{
		Threshold: 1,
		Addresses: []ids.ShortID{addr},
	}

	addressedCallPayload, err := message.NewRegisterL1Validator(
		subnetID,
		nodeID,
		blsPublicKey,
		expiry,
		balanceOwners,
		disableOwners,
		weight,
	)
	require.NoError(err)

	managerAddress := ids.ShortID{}
	addressedCall, err := payload.NewAddressedCall(
		managerAddress[:],
		addressedCallPayload.Bytes(),
	)
	require.NoError(err)

	unsignedWarpMessage, err := warp.NewUnsignedMessage(
		1,
		ids.GenerateTestID(),
		addressedCall.Bytes(),
	)
	require.NoError(err)

	emptySignature := &warp.BitSetSignature{
		Signers:   []byte{},
		Signature: [96]byte{},
	}

	signedWarpMessage, err := warp.NewMessage(
		unsignedWarpMessage,
		emptySignature,
	)
	require.NoError(err)

	testCases := []struct {
		name          string
		tx            platformvmtxs.UnsignedTx
		expectedChain constants.ChainAlias
	}{
		{
			name:          "BaseTx should be detected as P-Chain",
			tx:            &platformvmtxs.BaseTx{BaseTx: baseTxData.BaseTx},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "AdvanceTimeTx should be detected as P-Chain",
			tx: &platformvmtxs.AdvanceTimeTx{
				Time: 1234567890,
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "RewardValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.RewardValidatorTx{
				TxID: txID,
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "AddValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.AddValidatorTx{
				BaseTx:    baseTxData,
				StakeOuts: []*avax.TransferableOutput{},
				RewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "AddSubnetValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.AddSubnetValidatorTx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "AddDelegatorTx should be detected as P-Chain",
			tx: &platformvmtxs.AddDelegatorTx{
				BaseTx:    baseTxData,
				StakeOuts: []*avax.TransferableOutput{},
				DelegationRewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "RemoveSubnetValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.RemoveSubnetValidatorTx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "TransformSubnetTx should be detected as P-Chain",
			tx: &platformvmtxs.TransformSubnetTx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "AddPermissionlessValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.AddPermissionlessValidatorTx{
				BaseTx:    baseTxData,
				Signer:    &signer.Empty{},
				StakeOuts: []*avax.TransferableOutput{},
				ValidatorRewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
				DelegatorRewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "AddPermissionlessDelegatorTx should be detected as P-Chain",
			tx: &platformvmtxs.AddPermissionlessDelegatorTx{
				BaseTx:    baseTxData,
				StakeOuts: []*avax.TransferableOutput{},
				DelegationRewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "TransferSubnetOwnershipTx should be detected as P-Chain",
			tx: &platformvmtxs.TransferSubnetOwnershipTx{
				BaseTx: baseTxData,
				Subnet: ids.GenerateTestID(),
				Owner: &secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
				SubnetAuth: &secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "CreateSubnetTx should be detected as P-Chain",
			tx: &platformvmtxs.CreateSubnetTx{
				BaseTx: baseTxData,
				Owner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "CreateChainTx should be detected as P-Chain",
			tx: &platformvmtxs.CreateChainTx{
				BaseTx:    baseTxData,
				SubnetID:  ids.GenerateTestID(),
				ChainName: "test-chain",
				VMID:      ids.GenerateTestID(),
				SubnetAuth: &secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "ImportTx should be detected as P-Chain",
			tx: &platformvmtxs.ImportTx{
				BaseTx:      baseTxData,
				SourceChain: ids.GenerateTestID(),
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "ExportTx should be detected as P-Chain",
			tx: &platformvmtxs.ExportTx{
				BaseTx:           baseTxData,
				DestinationChain: ids.GenerateTestID(),
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "RegisterL1ValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.RegisterL1ValidatorTx{
				BaseTx:  baseTxData,
				Balance: 1000,
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "SetL1ValidatorWeightTx should be detected as P-Chain",
			tx: &platformvmtxs.SetL1ValidatorWeightTx{
				BaseTx: baseTxData,
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "IncreaseL1ValidatorBalanceTx should be detected as P-Chain",
			tx: &platformvmtxs.IncreaseL1ValidatorBalanceTx{
				BaseTx:       baseTxData,
				ValidationID: ids.GenerateTestID(),
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "DisableL1ValidatorTx should be detected as P-Chain",
			tx: &platformvmtxs.DisableL1ValidatorTx{
				BaseTx:       baseTxData,
				ValidationID: ids.GenerateTestID(),
				DisableAuth: &secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "ConvertSubnetToL1Tx should be detected as P-Chain",
			tx: &platformvmtxs.ConvertSubnetToL1Tx{
				BaseTx:     baseTxData,
				Subnet:     ids.GenerateTestID(),
				ChainID:    ids.GenerateTestID(),
				Address:    []byte{1, 2, 3, 4},
				Validators: validators,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name: "RegisterL1ValidatorTx with warp message should be detected as P-Chain",
			tx: &platformvmtxs.RegisterL1ValidatorTx{
				BaseTx:  baseTxData,
				Balance: 100000,
				Message: signedWarpMessage.Bytes(),
			},
			expectedChain: constants.PChainAlias,
		},
		{
			name:          "Invalid bytes should return UndefinedAlias",
			tx:            nil,
			expectedChain: constants.UndefinedAlias,
		},
		{
			name:          "Empty bytes should return UndefinedAlias",
			tx:            nil,
			expectedChain: constants.UndefinedAlias,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			var unsignedBytes []byte
			var err error

			// Handle special cases for invalid/empty bytes tests
			if tc.tx == nil {
				if i == len(testCases)-2 {
					// Invalid bytes test
					unsignedBytes = []byte{0xFF, 0xFF, 0xFF}
				} else {
					// Empty bytes test
					unsignedBytes = []byte{}
				}
			} else {
				unsignedBytes, err = platformvmtxs.Codec.Marshal(platformvmtxs.CodecVersion, &tc.tx)
				require.NoError(err)
			}

			result := AutoDetectChain(unsignedBytes)
			require.Equal(tc.expectedChain, result)
		})
	}
}

func TestAutoDetectChainXChainTxs(t *testing.T) {
	require := require.New(t)

	addr := ids.ShortID{
		0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
		0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
		0x44, 0x55, 0x66, 0x77,
	}

	txID := ids.ID{
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
	}

	avaxAssetID, err := ids.FromString("FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z")
	require.NoError(err)

	baseTxData := avax.BaseTx{
		NetworkID:    1,
		BlockchainID: ids.GenerateTestID(),
		Ins: []*avax.TransferableInput{
			{
				UTXOID: avax.UTXOID{
					TxID:        txID,
					OutputIndex: 1,
				},
				Asset: avax.Asset{
					ID: avaxAssetID,
				},
				In: &secp256k1fx.TransferInput{
					Amt: 1000000,
					Input: secp256k1fx.Input{
						SigIndices: []uint32{0},
					},
				},
			},
		},
		Outs: []*avax.TransferableOutput{
			{
				Asset: avax.Asset{
					ID: avaxAssetID,
				},
				Out: &secp256k1fx.TransferOutput{
					Amt: 500000,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			},
		},
	}

	testCases := []struct {
		name          string
		buildTx       func() (avmtxs.UnsignedTx, error)
		expectedChain constants.ChainAlias
	}{
		{
			name: "BaseTx should be detected as X-Chain",
			buildTx: func() (avmtxs.UnsignedTx, error) {
				return &avmtxs.BaseTx{
					BaseTx: baseTxData,
				}, nil
			},
			expectedChain: constants.XChainAlias,
		},
		{
			name: "CreateAssetTx should be detected as X-Chain",
			buildTx: func() (avmtxs.UnsignedTx, error) {
				transferOutput := &secp256k1fx.TransferOutput{
					Amt: 1000000,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				}
				return &avmtxs.CreateAssetTx{
					BaseTx:       avmtxs.BaseTx{BaseTx: baseTxData},
					Name:         "Test Asset",
					Symbol:       "TST",
					Denomination: 8,
					States: []*avmtxs.InitialState{
						{
							FxIndex: 0,
							Outs:    []verify.State{transferOutput},
						},
					},
				}, nil
			},
			expectedChain: constants.XChainAlias,
		},
		{
			name: "OperationTx should be detected as X-Chain",
			buildTx: func() (avmtxs.UnsignedTx, error) {
				return &avmtxs.OperationTx{
					BaseTx: avmtxs.BaseTx{BaseTx: baseTxData},
					Ops: []*avmtxs.Operation{
						{
							Asset: avax.Asset{
								ID: avaxAssetID,
							},
							UTXOIDs: []*avax.UTXOID{
								{
									TxID:        txID,
									OutputIndex: 0,
								},
							},
							Op: &secp256k1fx.MintOperation{
								MintInput: secp256k1fx.Input{
									SigIndices: []uint32{0},
								},
								MintOutput: secp256k1fx.MintOutput{
									OutputOwners: secp256k1fx.OutputOwners{
										Threshold: 1,
										Addrs:     []ids.ShortID{addr},
									},
								},
								TransferOutput: secp256k1fx.TransferOutput{
									Amt: 100000,
									OutputOwners: secp256k1fx.OutputOwners{
										Threshold: 1,
										Addrs:     []ids.ShortID{addr},
									},
								},
							},
						},
					},
				}, nil
			},
			expectedChain: constants.XChainAlias,
		},
		{
			name: "ImportTx should be detected as X-Chain",
			buildTx: func() (avmtxs.UnsignedTx, error) {
				return &avmtxs.ImportTx{
					BaseTx:      avmtxs.BaseTx{BaseTx: baseTxData},
					SourceChain: ids.GenerateTestID(),
					ImportedIns: []*avax.TransferableInput{
						{
							UTXOID: avax.UTXOID{
								TxID:        txID,
								OutputIndex: 2,
							},
							Asset: avax.Asset{
								ID: avaxAssetID,
							},
							In: &secp256k1fx.TransferInput{
								Amt: 2000000,
								Input: secp256k1fx.Input{
									SigIndices: []uint32{0},
								},
							},
						},
					},
				}, nil
			},
			expectedChain: constants.XChainAlias,
		},
		{
			name: "ExportTx should be detected as X-Chain",
			buildTx: func() (avmtxs.UnsignedTx, error) {
				return &avmtxs.ExportTx{
					BaseTx:           avmtxs.BaseTx{BaseTx: baseTxData},
					DestinationChain: ids.GenerateTestID(),
					ExportedOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: avaxAssetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 300000,
								OutputOwners: secp256k1fx.OutputOwners{
									Threshold: 1,
									Addrs:     []ids.ShortID{addr},
								},
							},
						},
					},
				}, nil
			},
			expectedChain: constants.XChainAlias,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			unsignedTx, err := tc.buildTx()
			require.NoError(err)

			// Marshal the unsigned transaction to bytes
			unsignedBytes, err := builder.Parser.Codec().Marshal(avmtxs.CodecVersion, &unsignedTx)
			require.NoError(err)

			result := AutoDetectChain(unsignedBytes)
			require.Equal(tc.expectedChain, result)
		})
	}
}

func TestAutoDetectChainCChainTxs(t *testing.T) {
	require := require.New(t)

	addr := ids.ShortID{
		0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
		0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
		0x44, 0x55, 0x66, 0x77,
	}

	txID := ids.ID{
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
	}

	avaxAssetID, err := ids.FromString("FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z")
	require.NoError(err)

	ethAddr := common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	testCases := []struct {
		name          string
		buildTx       func() (atomic.UnsignedAtomicTx, error)
		expectedChain constants.ChainAlias
	}{
		{
			name: "ImportTx should be detected as C-Chain",
			buildTx: func() (atomic.UnsignedAtomicTx, error) {
				return &atomic.UnsignedImportTx{
					NetworkID:    1,
					BlockchainID: ids.GenerateTestID(),
					SourceChain:  ids.GenerateTestID(),
					ImportedInputs: []*avax.TransferableInput{
						{
							UTXOID: avax.UTXOID{
								TxID:        txID,
								OutputIndex: 0,
							},
							Asset: avax.Asset{
								ID: avaxAssetID,
							},
							In: &secp256k1fx.TransferInput{
								Amt: 1000000,
								Input: secp256k1fx.Input{
									SigIndices: []uint32{0},
								},
							},
						},
					},
					Outs: []atomic.EVMOutput{
						{
							Address: ethAddr,
							Amount:  500000,
							AssetID: avaxAssetID,
						},
					},
				}, nil
			},
			expectedChain: constants.CChainAlias,
		},
		{
			name: "ExportTx should be detected as C-Chain",
			buildTx: func() (atomic.UnsignedAtomicTx, error) {
				return &atomic.UnsignedExportTx{
					NetworkID:        1,
					BlockchainID:     ids.GenerateTestID(),
					DestinationChain: ids.GenerateTestID(),
					Ins: []atomic.EVMInput{
						{
							Address: ethAddr,
							Amount:  1000000,
							AssetID: avaxAssetID,
							Nonce:   0,
						},
					},
					ExportedOutputs: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: avaxAssetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 500000,
								OutputOwners: secp256k1fx.OutputOwners{
									Threshold: 1,
									Addrs:     []ids.ShortID{addr},
								},
							},
						},
					},
				}, nil
			},
			expectedChain: constants.CChainAlias,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			unsignedTx, err := tc.buildTx()
			require.NoError(err)

			// Marshal the unsigned transaction to bytes
			unsignedBytes, err := atomic.Codec.Marshal(atomic.CodecVersion, &unsignedTx)
			require.NoError(err)

			result := AutoDetectChain(unsignedBytes)
			require.Equal(tc.expectedChain, result)
		})
	}
}

func TestGetNetworkIDPChainTxs(t *testing.T) {
	require := require.New(t)
	networkID := uint32(1337)
	addr := ids.ShortEmpty

	baseTxData := platformvmtxs.BaseTx{
		BaseTx: avax.BaseTx{
			NetworkID:    networkID,
			BlockchainID: ids.Empty,
		},
	}

	testCases := []struct {
		name        string
		createTx    func() platformvmtxs.UnsignedTx
		expectedNet uint32
	}{
		{
			name: "AddValidatorTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.AddValidatorTx{
					BaseTx:    baseTxData,
					StakeOuts: []*avax.TransferableOutput{},
					RewardsOwner: &secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "AddDelegatorTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.AddDelegatorTx{
					BaseTx:    baseTxData,
					StakeOuts: []*avax.TransferableOutput{},
					DelegationRewardsOwner: &secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "AddPermissionlessValidatorTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.AddPermissionlessValidatorTx{
					BaseTx: baseTxData,
					Signer: &signer.Empty{},
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: ids.Empty,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 500000,
								OutputOwners: secp256k1fx.OutputOwners{
									Threshold: 1,
									Addrs:     []ids.ShortID{addr},
								},
							},
						},
					},
					ValidatorRewardsOwner: &secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
					DelegatorRewardsOwner: &secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "AddPermissionlessDelegatorTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.AddPermissionlessDelegatorTx{
					BaseTx: baseTxData,
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: ids.Empty,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 500000,
								OutputOwners: secp256k1fx.OutputOwners{
									Threshold: 1,
									Addrs:     []ids.ShortID{addr},
								},
							},
						},
					},
					DelegationRewardsOwner: &secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "CreateSubnetTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.CreateSubnetTx{
					BaseTx: baseTxData,
					Owner: &secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "CreateChainTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.CreateChainTx{
					BaseTx:    baseTxData,
					SubnetID:  ids.Empty,
					ChainName: "testchain",
					VMID:      ids.Empty,
					SubnetAuth: &secp256k1fx.Input{
						SigIndices: []uint32{0},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "ImportTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.ImportTx{
					BaseTx:         baseTxData,
					SourceChain:    ids.Empty,
					ImportedInputs: []*avax.TransferableInput{},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "ExportTx",
			createTx: func() platformvmtxs.UnsignedTx {
				return &platformvmtxs.ExportTx{
					BaseTx:           baseTxData,
					DestinationChain: ids.Empty,
					ExportedOutputs: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: ids.Empty,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 500000,
								OutputOwners: secp256k1fx.OutputOwners{
									Threshold: 1,
									Addrs:     []ids.ShortID{addr},
								},
							},
						},
					},
				}
			},
			expectedNet: networkID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			unsignedTx := tc.createTx()

			// Marshal the unsigned transaction to bytes
			unsignedBytes, err := platformvmtxs.Codec.Marshal(platformvmtxs.CodecVersion, &unsignedTx)
			require.NoError(err)

			// Test GetNetworkID (auto-detect)
			result := GetNetworkID(unsignedBytes)
			require.Equal(tc.expectedNet, result)

			// Test GetPChainTxNetworkID (direct)
			result = GetPChainTxNetworkID(unsignedTx)
			require.Equal(tc.expectedNet, result)
		})
	}
}

func TestGetNetworkIDXChainTxs(t *testing.T) {
	require := require.New(t)
	networkID := uint32(1337)

	testCases := []struct {
		name        string
		createTx    func() avmtxs.UnsignedTx
		expectedNet uint32
	}{
		{
			name: "BaseTx",
			createTx: func() avmtxs.UnsignedTx {
				return &avmtxs.BaseTx{
					BaseTx: avax.BaseTx{
						NetworkID: networkID,
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "CreateAssetTx",
			createTx: func() avmtxs.UnsignedTx {
				return &avmtxs.CreateAssetTx{
					BaseTx: avmtxs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: networkID,
						},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "OperationTx",
			createTx: func() avmtxs.UnsignedTx {
				return &avmtxs.OperationTx{
					BaseTx: avmtxs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: networkID,
						},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "ImportTx",
			createTx: func() avmtxs.UnsignedTx {
				return &avmtxs.ImportTx{
					BaseTx: avmtxs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: networkID,
						},
					},
				}
			},
			expectedNet: networkID,
		},
		{
			name: "ExportTx",
			createTx: func() avmtxs.UnsignedTx {
				return &avmtxs.ExportTx{
					BaseTx: avmtxs.BaseTx{
						BaseTx: avax.BaseTx{
							NetworkID: networkID,
						},
					},
				}
			},
			expectedNet: networkID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			unsignedTx := tc.createTx()

			// Marshal the unsigned transaction to bytes
			unsignedBytes, err := builder.Parser.Codec().Marshal(avmtxs.CodecVersion, &unsignedTx)
			require.NoError(err)

			// Test GetNetworkID (auto-detect)
			result := GetNetworkID(unsignedBytes)
			require.Equal(tc.expectedNet, result)

			// Test GetXChainTxNetworkID (direct)
			result = GetXChainTxNetworkID(unsignedTx)
			require.Equal(tc.expectedNet, result)
		})
	}
}

func TestGetNetworkIDCChainTxs(t *testing.T) {
	require := require.New(t)
	networkID := uint32(1337)

	testCases := []struct {
		name        string
		createTx    func() atomic.UnsignedAtomicTx
		expectedNet uint32
	}{
		{
			name: "UnsignedImportTx",
			createTx: func() atomic.UnsignedAtomicTx {
				return &atomic.UnsignedImportTx{
					NetworkID: networkID,
				}
			},
			expectedNet: networkID,
		},
		{
			name: "UnsignedExportTx",
			createTx: func() atomic.UnsignedAtomicTx {
				return &atomic.UnsignedExportTx{
					NetworkID: networkID,
				}
			},
			expectedNet: networkID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			unsignedTx := tc.createTx()

			// Marshal the unsigned transaction to bytes
			unsignedBytes, err := atomic.Codec.Marshal(atomic.CodecVersion, &unsignedTx)
			require.NoError(err)

			// Test GetNetworkID (auto-detect)
			result := GetNetworkID(unsignedBytes)
			require.Equal(tc.expectedNet, result)

			// Test GetCChainTxNetworkID (direct)
			result = GetCChainTxNetworkID(unsignedTx)
			require.Equal(tc.expectedNet, result)
		})
	}
}

func TestGetNetworkIDInvalidTxs(t *testing.T) {
	require := require.New(t)

	testCases := []struct {
		name        string
		txBytes     []byte
		expectedNet uint32
	}{
		{
			name:        "Empty bytes",
			txBytes:     []byte{},
			expectedNet: 0,
		},
		{
			name:        "Invalid bytes",
			txBytes:     []byte{0x00, 0x01, 0x02},
			expectedNet: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			result := GetNetworkID(tc.txBytes)
			require.Equal(tc.expectedNet, result)
		})
	}
}
