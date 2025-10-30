// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ledger

import (
	"errors"
	"testing"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger/ledgermock"
)

var errTest = errors.New("test")

func TestNewLedgerKeychain(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	pubKey, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	// user request invalid number of addresses to derive
	ledger := ledgermock.NewLedger(ctrl)
	_, err = NewKeychain(ledger, 0)
	require.ErrorIs(err, ErrInvalidNumAddrsToDerive)

	// ledger does not return expected number of derived addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{}, nil).Times(1)
	_, err = NewKeychain(ledger, 1)
	require.ErrorIs(err, ErrInvalidNumAddrsDerived)

	// ledger return error when asked for derived addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey.PublicKey()}, errTest).Times(1)
	_, err = NewKeychain(ledger, 1)
	require.ErrorIs(err, errTest)

	// good path
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey.PublicKey()}, nil).Times(1)
	_, err = NewKeychain(ledger, 1)
	require.NoError(err)
}

func TestLedgerKeychain_Addresses(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key1, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key2, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key3, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	pubKey1 := key1.PublicKey()
	pubKey2 := key2.PublicKey()
	pubKey3 := key3.PublicKey()

	// 1 addr
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	addrs := kc.Addresses()
	require.Len(addrs, 1)
	require.True(addrs.Contains(pubKey1.Address()))

	// multiple addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0, 1, 2}).Return([]*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}, nil).Times(1)
	kc, err = NewKeychain(ledger, 3)
	require.NoError(err)

	addrs = kc.Addresses()
	require.Len(addrs, 3)
	require.Contains(addrs, pubKey1.Address())
	require.Contains(addrs, pubKey2.Address())
	require.Contains(addrs, pubKey3.Address())
}

func TestLedgerKeychain_Get(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key1, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key2, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key3, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	pubKey1 := key1.PublicKey()
	pubKey2 := key2.PublicKey()
	pubKey3 := key3.PublicKey()

	addr1 := pubKey1.Address()
	addr2 := pubKey2.Address()
	addr3 := pubKey3.Address()

	// 1 addr
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	_, b := kc.Get(ids.GenerateTestShortID())
	require.False(b)

	s, b := kc.Get(addr1)
	require.Equal(s.Address(), addr1)
	require.True(b)

	// multiple addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0, 1, 2}).Return([]*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}, nil).Times(1)
	kc, err = NewKeychain(ledger, 3)
	require.NoError(err)

	_, b = kc.Get(ids.GenerateTestShortID())
	require.False(b)

	s, b = kc.Get(addr1)
	require.True(b)
	require.Equal(s.Address(), addr1)

	s, b = kc.Get(addr2)
	require.True(b)
	require.Equal(s.Address(), addr2)

	s, b = kc.Get(addr3)
	require.True(b)
	require.Equal(s.Address(), addr3)
}

func TestLedgerSigner_SignHash(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key1, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key2, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key3, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	pubKey1 := key1.PublicKey()
	pubKey2 := key2.PublicKey()
	pubKey3 := key3.PublicKey()

	addr1 := pubKey1.Address()
	addr2 := pubKey2.Address()
	addr3 := pubKey3.Address()

	toSign := []byte{1, 2, 3, 4, 5}
	expectedSignature1 := []byte{1, 1, 1}
	expectedSignature2 := []byte{2, 2, 2}
	expectedSignature3 := []byte{3, 3, 3}

	// ledger returns an incorrect number of signatures
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{}, nil).Times(1)
	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	s, b := kc.Get(addr1)
	require.True(b)

	_, err = s.SignHash(toSign)
	require.ErrorIs(err, ErrInvalidNumSignatures)

	// ledger returns an error when asked for signature
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{expectedSignature1}, errTest).Times(1)
	kc, err = NewKeychain(ledger, 1)
	require.NoError(err)

	s, b = kc.Get(addr1)
	require.True(b)

	_, err = s.SignHash(toSign)
	require.ErrorIs(err, errTest)

	// good path 1 addr
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{expectedSignature1}, nil).Times(1)
	kc, err = NewKeychain(ledger, 1)
	require.NoError(err)

	s, b = kc.Get(addr1)
	require.True(b)

	signature, err := s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature1, signature)

	// good path 3 addr
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0, 1, 2}).Return([]*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{expectedSignature1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{1}).Return([][]byte{expectedSignature2}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{2}).Return([][]byte{expectedSignature3}, nil).Times(1)
	kc, err = NewKeychain(ledger, 3)
	require.NoError(err)

	s, b = kc.Get(addr1)
	require.True(b)

	signature, err = s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature1, signature)

	s, b = kc.Get(addr2)
	require.True(b)

	signature, err = s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature2, signature)

	s, b = kc.Get(addr3)
	require.True(b)

	signature, err = s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature3, signature)
}

func TestNewLedgerKeychainFromIndices(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	pubKey := key.PublicKey()

	// user request invalid number of indices
	ledger := ledgermock.NewLedger(ctrl)
	_, err = NewKeychainFromIndices(ledger, []uint32{})
	require.ErrorIs(err, ErrInvalidIndicesLength)

	// ledger does not return expected number of derived addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{}, nil).Times(1)
	_, err = NewKeychainFromIndices(ledger, []uint32{0})
	require.ErrorIs(err, ErrInvalidNumAddrsDerived)

	// ledger return error when asked for derived addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey}, errTest).Times(1)
	_, err = NewKeychainFromIndices(ledger, []uint32{0})
	require.ErrorIs(err, errTest)

	// good path
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey}, nil).Times(1)
	_, err = NewKeychainFromIndices(ledger, []uint32{0})
	require.NoError(err)
}

func TestLedgerKeychainFromIndices_Addresses(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key1, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key2, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key3, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	pubKey1 := key1.PublicKey()
	pubKey2 := key2.PublicKey()
	pubKey3 := key3.PublicKey()

	addr1 := pubKey1.Address()
	addr2 := pubKey2.Address()
	addr3 := pubKey3.Address()

	// 1 addr
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	kc, err := NewKeychainFromIndices(ledger, []uint32{0})
	require.NoError(err)

	addrs := kc.Addresses()
	require.Len(addrs, 1)
	require.True(addrs.Contains(addr1))

	// first 3 addresses
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0, 1, 2}).Return([]*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}, nil).Times(1)
	kc, err = NewKeychainFromIndices(ledger, []uint32{0, 1, 2})
	require.NoError(err)

	addrs = kc.Addresses()
	require.Len(addrs, 3)
	require.Contains(addrs, addr1)
	require.Contains(addrs, addr2)
	require.Contains(addrs, addr3)

	// some 3 addresses
	indices := []uint32{3, 7, 1}
	pubKeys := []*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys(indices).Return(pubKeys, nil).Times(1)
	kc, err = NewKeychainFromIndices(ledger, indices)
	require.NoError(err)

	addrs = kc.Addresses()
	require.Len(addrs, len(indices))
	require.Contains(addrs, addr1)
	require.Contains(addrs, addr2)
	require.Contains(addrs, addr3)

	// repeated addresses
	indices = []uint32{3, 7, 1, 3, 1, 7}
	pubKeys = []*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3, pubKey1, pubKey2, pubKey3}
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys(indices).Return(pubKeys, nil).Times(1)
	kc, err = NewKeychainFromIndices(ledger, indices)
	require.NoError(err)

	addrs = kc.Addresses()
	require.Len(addrs, 3)
	require.Contains(addrs, addr1)
	require.Contains(addrs, addr2)
	require.Contains(addrs, addr3)
}

func TestLedgerKeychainFromIndices_Get(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key1, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key2, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key3, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	pubKey1 := key1.PublicKey()
	pubKey2 := key2.PublicKey()
	pubKey3 := key3.PublicKey()

	addr1 := pubKey1.Address()
	addr2 := pubKey2.Address()
	addr3 := pubKey3.Address()

	// 1 addr
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	kc, err := NewKeychainFromIndices(ledger, []uint32{0})
	require.NoError(err)

	_, b := kc.Get(ids.GenerateTestShortID())
	require.False(b)

	s, b := kc.Get(addr1)
	require.Equal(s.Address(), addr1)
	require.True(b)

	// some 3 addresses
	indices := []uint32{3, 7, 1}
	pubKeys := []*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys(indices).Return(pubKeys, nil).Times(1)
	kc, err = NewKeychainFromIndices(ledger, indices)
	require.NoError(err)

	_, b = kc.Get(ids.GenerateTestShortID())
	require.False(b)

	s, b = kc.Get(addr1)
	require.True(b)
	require.Equal(s.Address(), addr1)

	s, b = kc.Get(addr2)
	require.True(b)
	require.Equal(s.Address(), addr2)

	s, b = kc.Get(addr3)
	require.True(b)
	require.Equal(s.Address(), addr3)
}

func TestLedgerSignerFromIndices_SignHash(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key1, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key2, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	key3, err := secp256k1.NewPrivateKey()
	require.NoError(err)

	pubKey1 := key1.PublicKey()
	pubKey2 := key2.PublicKey()
	pubKey3 := key3.PublicKey()

	addr1 := pubKey1.Address()
	addr2 := pubKey2.Address()
	addr3 := pubKey3.Address()

	toSign := []byte{1, 2, 3, 4, 5}
	expectedSignature1 := []byte{1, 1, 1}
	expectedSignature2 := []byte{2, 2, 2}
	expectedSignature3 := []byte{3, 3, 3}

	// ledger returns an incorrect number of signatures
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{}, nil).Times(1)
	kc, err := NewKeychainFromIndices(ledger, []uint32{0})
	require.NoError(err)

	s, b := kc.Get(addr1)
	require.True(b)

	_, err = s.SignHash(toSign)
	require.ErrorIs(err, ErrInvalidNumSignatures)

	// ledger returns an error when asked for signature
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{expectedSignature1}, errTest).Times(1)
	kc, err = NewKeychainFromIndices(ledger, []uint32{0})
	require.NoError(err)

	s, b = kc.Get(addr1)
	require.True(b)

	_, err = s.SignHash(toSign)
	require.ErrorIs(err, errTest)

	// good path 1 addr
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{0}).Return([][]byte{expectedSignature1}, nil).Times(1)
	kc, err = NewKeychainFromIndices(ledger, []uint32{0})
	require.NoError(err)

	s, b = kc.Get(addr1)
	require.True(b)

	signature, err := s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature1, signature)

	// good path some 3 addresses
	indices := []uint32{3, 7, 1}
	pubKeys := []*secp256k1.PublicKey{pubKey1, pubKey2, pubKey3}
	ledger = ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys(indices).Return(pubKeys, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{indices[0]}).Return([][]byte{expectedSignature1}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{indices[1]}).Return([][]byte{expectedSignature2}, nil).Times(1)
	ledger.EXPECT().SignHash(toSign, []uint32{indices[2]}).Return([][]byte{expectedSignature3}, nil).Times(1)
	kc, err = NewKeychainFromIndices(ledger, indices)
	require.NoError(err)

	s, b = kc.Get(addr1)
	require.True(b)

	signature, err = s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature1, signature)

	s, b = kc.Get(addr2)
	require.True(b)

	signature, err = s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature2, signature)

	s, b = kc.Get(addr3)
	require.True(b)

	signature, err = s.SignHash(toSign)
	require.NoError(err)
	require.Equal(expectedSignature3, signature)
}

func TestShouldUseSignHash(t *testing.T) {
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

	baseTxData := txs.BaseTx{
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

	testCases := []struct {
		name     string
		tx       txs.UnsignedTx
		expected bool
	}{
		{
			name: "TransferSubnetOwnershipTx should use SignHash",
			tx: &txs.TransferSubnetOwnershipTx{
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
			expected: true,
		},
		{
			name:     "BaseTx should not use SignHash",
			tx:       &txs.BaseTx{BaseTx: baseTxData.BaseTx},
			expected: false,
		},
		{
			name: "AddValidatorTx should not use SignHash",
			tx: &txs.AddValidatorTx{
				BaseTx:    baseTxData,
				StakeOuts: []*avax.TransferableOutput{},
				RewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expected: false,
		},
		{
			name: "AddSubnetValidatorTx should not use SignHash",
			tx: &txs.AddSubnetValidatorTx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expected: false,
		},
		{
			name: "AddDelegatorTx should not use SignHash",
			tx: &txs.AddDelegatorTx{
				BaseTx:    baseTxData,
				StakeOuts: []*avax.TransferableOutput{},
				DelegationRewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expected: false,
		},
		{
			name: "RemoveSubnetValidatorTx should not use SignHash",
			tx: &txs.RemoveSubnetValidatorTx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expected: false,
		},
		{
			name: "TransformSubnetTx should not use SignHash",
			tx: &txs.TransformSubnetTx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expected: false,
		},
		{
			name: "AddPermissionlessValidatorTx should not use SignHash",
			tx: &txs.AddPermissionlessValidatorTx{
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
			expected: false,
		},
		{
			name: "AddPermissionlessDelegatorTx should not use SignHash",
			tx: &txs.AddPermissionlessDelegatorTx{
				BaseTx:    baseTxData,
				StakeOuts: []*avax.TransferableOutput{},
				DelegationRewardsOwner: &secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
			expected: false,
		},
		{
			name: "ConvertSubnetToL1Tx should not use SignHash",
			tx: &txs.ConvertSubnetToL1Tx{
				BaseTx:     baseTxData,
				SubnetAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
				Validators: []*txs.ConvertSubnetToL1Validator{},
			},
			expected: false,
		},
		{
			name:     "RegisterL1ValidatorTx should not use SignHash",
			tx:       &txs.RegisterL1ValidatorTx{BaseTx: baseTxData},
			expected: false,
		},
		{
			name:     "SetL1ValidatorWeightTx should not use SignHash",
			tx:       &txs.SetL1ValidatorWeightTx{BaseTx: baseTxData},
			expected: false,
		},
		{
			name:     "IncreaseL1ValidatorBalanceTx should not use SignHash",
			tx:       &txs.IncreaseL1ValidatorBalanceTx{BaseTx: baseTxData},
			expected: false,
		},
		{
			name: "DisableL1ValidatorTx should not use SignHash",
			tx: &txs.DisableL1ValidatorTx{
				BaseTx:      baseTxData,
				DisableAuth: &secp256k1fx.Input{SigIndices: []uint32{0}},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			unsignedBytes, err := txs.Codec.Marshal(txs.CodecVersion, &tc.tx)
			require.NoError(err)
			result, err := shouldUseSignHash(unsignedBytes)
			require.NoError(err)
			require.Equal(tc.expected, result)
		})
	}

	// Test invalid bytes - should return error
	t.Run("Invalid bytes should return error", func(_ *testing.T) {
		_, err := shouldUseSignHash([]byte{0xFF, 0xFF, 0xFF})
		require.ErrorIs(err, codec.ErrUnknownVersion)
	})
}

func TestLedgerSigner_Sign_WithPChainTransferSubnetOwnership(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	pubKey := key.PublicKey()
	addr := pubKey.Address()
	expectedSignature := []byte{1, 1, 1}

	// Create a TransferSubnetOwnershipTx
	baseTxData := txs.BaseTx{
		BaseTx: avax.BaseTx{
			NetworkID:    1,
			BlockchainID: ids.GenerateTestID(),
			Ins:          []*avax.TransferableInput{},
			Outs:         []*avax.TransferableOutput{},
		},
	}

	var tx txs.UnsignedTx = &txs.TransferSubnetOwnershipTx{
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
	}

	unsignedBytes, err := txs.Codec.Marshal(txs.CodecVersion, &tx)
	require.NoError(err)

	// When signing with P-Chain alias, should use SignHash
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey}, nil).Times(1)
	ledger.EXPECT().SignHash(hashing.ComputeHash256(unsignedBytes), []uint32{0}).Return([][]byte{expectedSignature}, nil).Times(1)

	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	s, b := kc.Get(addr)
	require.True(b)

	signature, err := s.Sign(unsignedBytes)
	require.NoError(err)
	require.Equal(expectedSignature, signature)
}

func TestLedgerSigner_Sign_WithPChainNonTransferSubnetOwnership(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	pubKey := key.PublicKey()
	addr := pubKey.Address()
	expectedSignature := []byte{2, 2, 2}

	// Create a BaseTx
	var tx txs.UnsignedTx = &txs.BaseTx{
		BaseTx: avax.BaseTx{
			NetworkID:    1,
			BlockchainID: ids.GenerateTestID(),
			Ins:          []*avax.TransferableInput{},
			Outs:         []*avax.TransferableOutput{},
		},
	}

	unsignedBytes, err := txs.Codec.Marshal(txs.CodecVersion, &tx)
	require.NoError(err)

	// When signing with P-Chain alias but NOT TransferSubnetOwnershipTx, should use Sign (not SignHash)
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey}, nil).Times(1)
	ledger.EXPECT().Sign(unsignedBytes, []uint32{0}).Return([][]byte{expectedSignature}, nil).Times(1)

	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	s, b := kc.Get(addr)
	require.True(b)

	signature, err := s.Sign(unsignedBytes)
	require.NoError(err)
	require.Equal(expectedSignature, signature)
}

func TestLedgerSigner_Sign_WithoutChainAlias(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	pubKey := key.PublicKey()
	addr := pubKey.Address()
	toSign := []byte{1, 2, 3, 4, 5}
	expectedSignature := []byte{3, 3, 3}

	// When signing without chain alias, should use Sign (not SignHash)
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey}, nil).Times(1)
	ledger.EXPECT().Sign(toSign, []uint32{0}).Return([][]byte{expectedSignature}, nil).Times(1)

	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	s, b := kc.Get(addr)
	require.True(b)

	signature, err := s.Sign(toSign)
	require.NoError(err)
	require.Equal(expectedSignature, signature)
}

func TestLedgerSigner_Sign_WithNonPChain(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	key, err := secp256k1.NewPrivateKey()
	require.NoError(err)
	pubKey := key.PublicKey()
	addr := pubKey.Address()
	toSign := []byte{1, 2, 3, 4, 5}
	expectedSignature := []byte{4, 4, 4}

	// When signing with non-P-Chain alias, should use Sign (not SignHash)
	ledger := ledgermock.NewLedger(ctrl)
	ledger.EXPECT().PubKeys([]uint32{0}).Return([]*secp256k1.PublicKey{pubKey}, nil).Times(1)
	ledger.EXPECT().Sign(toSign, []uint32{0}).Return([][]byte{expectedSignature}, nil).Times(1)

	kc, err := NewKeychain(ledger, 1)
	require.NoError(err)

	s, b := kc.Get(addr)
	require.True(b)

	signature, err := s.Sign(toSign)
	require.NoError(err)
	require.Equal(expectedSignature, signature)
}
