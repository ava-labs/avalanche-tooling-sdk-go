// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mocks "github.com/ava-labs/avalanche-tooling-sdk-go/mocks/keychain"
)

var errTest = errors.New("test error")

func TestNewSigner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	t.Run("success", func(t *testing.T) {
		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)
		require.NotNil(t, signer)
		require.Equal(t, testAddr, signer.Address())
		require.False(t, signer.IsNoOp())
	})

	t.Run("nil keychain", func(t *testing.T) {
		signer, err := NewSigner(nil)
		require.Error(t, err)
		require.Nil(t, signer)
		require.Contains(t, err.Error(), "keychain cannot be nil")
	})

	t.Run("no addresses in keychain", func(t *testing.T) {
		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		addrSet := set.Set[common.Address]{}

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)

		signer, err := NewSigner(mockKeychain)
		require.Error(t, err)
		require.Nil(t, signer)
		require.Contains(t, err.Error(), "expected keychain to have 1 address, found 0")
	})

	t.Run("multiple addresses in keychain", func(t *testing.T) {
		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)
		addrSet.Add(common.HexToAddress("0x9876543210987654321098765432109876543210"))

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)

		signer, err := NewSigner(mockKeychain)
		require.Error(t, err)
		require.Nil(t, signer)
		require.Contains(t, err.Error(), "expected keychain to have 1 address, found 2")
	})

	t.Run("GetEth returns false", func(t *testing.T) {
		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(nil, false)

		signer, err := NewSigner(mockKeychain)
		require.Error(t, err)
		require.Nil(t, signer)
		require.Contains(t, err.Error(), "unexpected failure obtaining unique signer from keychain")
	})
}

func TestNewNoOpSigner(t *testing.T) {
	testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	signer := NewNoOpSigner(testAddr)
	require.NotNil(t, signer)
	require.Equal(t, testAddr, signer.Address())
	require.True(t, signer.IsNoOp())
}

func TestSigner_IsNoOp(t *testing.T) {
	t.Run("nil signer", func(t *testing.T) {
		var signer *Signer
		require.True(t, signer.IsNoOp())
	})

	t.Run("NoOp signer", func(t *testing.T) {
		signer := NewNoOpSigner(common.HexToAddress("0x1234567890123456789012345678901234567890"))
		require.True(t, signer.IsNoOp())
	})

	t.Run("crypto signer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)
		require.False(t, signer.IsNoOp())
	})
}

func TestSigner_Address(t *testing.T) {
	t.Run("nil signer", func(t *testing.T) {
		var signer *Signer
		addr := signer.Address()
		require.Equal(t, common.Address{}, addr)
	})

	t.Run("NoOp signer", func(t *testing.T) {
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		signer := NewNoOpSigner(testAddr)
		require.Equal(t, testAddr, signer.Address())
	})

	t.Run("crypto signer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)
		require.Equal(t, testAddr, signer.Address())
	})
}

func TestSigner_SignTx(t *testing.T) {
	chainID := big.NewInt(43114)
	testTx := types.NewTransaction(0, common.HexToAddress("0x1"), big.NewInt(100), 21000, big.NewInt(1000000000), nil)

	t.Run("nil signer", func(t *testing.T) {
		var signer *Signer
		signedTx, err := signer.SignTx(chainID, testTx)
		require.Error(t, err)
		require.Nil(t, signedTx)
		require.Contains(t, err.Error(), "signer is nil")
	})

	t.Run("nil chainID", func(t *testing.T) {
		signer := NewNoOpSigner(common.HexToAddress("0x1234567890123456789012345678901234567890"))
		signedTx, err := signer.SignTx(nil, testTx)
		require.Error(t, err)
		require.Nil(t, signedTx)
		require.Contains(t, err.Error(), "chainID cannot be nil")
	})

	t.Run("nil transaction", func(t *testing.T) {
		signer := NewNoOpSigner(common.HexToAddress("0x1234567890123456789012345678901234567890"))
		signedTx, err := signer.SignTx(chainID, nil)
		require.Error(t, err)
		require.Nil(t, signedTx)
		require.Contains(t, err.Error(), "transaction cannot be nil")
	})

	t.Run("NoOp signer returns unsigned tx", func(t *testing.T) {
		signer := NewNoOpSigner(common.HexToAddress("0x1234567890123456789012345678901234567890"))
		signedTx, err := signer.SignTx(chainID, testTx)
		require.NoError(t, err)
		require.Equal(t, testTx, signedTx)
	})

	t.Run("crypto signer signs transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)

		// Create expected signature (65 bytes with valid v, r, s values)
		mockSignature := make([]byte, 65)
		mockSignature[64] = 0 // v value

		txSigner := types.LatestSignerForChainID(chainID)
		hash := txSigner.Hash(testTx)

		mockSigner.EXPECT().SignHash(hash.Bytes()).Return(mockSignature, nil)

		signedTx, err := signer.SignTx(chainID, testTx)
		require.NoError(t, err)
		require.NotNil(t, signedTx)
		require.NotEqual(t, testTx, signedTx)
	})

	t.Run("crypto signer signing fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)

		txSigner := types.LatestSignerForChainID(chainID)
		hash := txSigner.Hash(testTx)

		mockSigner.EXPECT().SignHash(hash.Bytes()).Return(nil, errTest)

		signedTx, err := signer.SignTx(chainID, testTx)
		require.Error(t, err)
		require.Nil(t, signedTx)
		require.Contains(t, err.Error(), "failed to sign transaction hash")
	})
}

func TestSigner_TransactOpts(t *testing.T) {
	chainID := big.NewInt(43114)

	t.Run("nil signer", func(t *testing.T) {
		var signer *Signer
		opts, err := signer.TransactOpts(chainID)
		require.Error(t, err)
		require.Nil(t, opts)
		require.Contains(t, err.Error(), "signer is nil")
	})

	t.Run("nil chainID", func(t *testing.T) {
		signer := NewNoOpSigner(common.HexToAddress("0x1234567890123456789012345678901234567890"))
		opts, err := signer.TransactOpts(nil)
		require.Error(t, err)
		require.Nil(t, opts)
		require.Contains(t, err.Error(), "chainID cannot be nil")
	})

	t.Run("NoOp signer sets NoSend", func(t *testing.T) {
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		signer := NewNoOpSigner(testAddr)
		opts, err := signer.TransactOpts(chainID)
		require.NoError(t, err)
		require.NotNil(t, opts)
		require.Equal(t, testAddr, opts.From)
		require.True(t, opts.NoSend)
		require.NotNil(t, opts.Signer)
		require.NotNil(t, opts.Context)
	})

	t.Run("crypto signer doesn't set NoSend", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)

		opts, err := signer.TransactOpts(chainID)
		require.NoError(t, err)
		require.NotNil(t, opts)
		require.Equal(t, testAddr, opts.From)
		require.False(t, opts.NoSend)
		require.NotNil(t, opts.Signer)
		require.NotNil(t, opts.Context)
	})

	t.Run("signer function validates address", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockKeychain := mocks.NewMockEthKeychain(ctrl)
		mockSigner := mocks.NewMockSigner(ctrl)
		testAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		wrongAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")

		addrSet := set.Set[common.Address]{}
		addrSet.Add(testAddr)

		mockKeychain.EXPECT().EthAddresses().Return(addrSet)
		mockKeychain.EXPECT().GetEth(testAddr).Return(mockSigner, true)

		signer, err := NewSigner(mockKeychain)
		require.NoError(t, err)

		opts, err := signer.TransactOpts(chainID)
		require.NoError(t, err)

		testTx := types.NewTransaction(0, common.HexToAddress("0x1"), big.NewInt(100), 21000, big.NewInt(1000000000), nil)

		_, err = opts.Signer(wrongAddr, testTx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address mismatch")
	})
}

func TestNewSignerFromPrivateKey(t *testing.T) {
	// Using the well-known EWOQ private key for testing
	const ewoqPrivateKey = "PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"

	t.Run("success", func(t *testing.T) {
		signer, err := NewSignerFromPrivateKey(ewoqPrivateKey)
		require.NoError(t, err)
		require.NotNil(t, signer)
		require.False(t, signer.IsNoOp())

		// Verify the address is correct for the EWOQ key
		expectedAddr := common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
		require.Equal(t, expectedAddr, signer.Address())
	})

	t.Run("invalid private key", func(t *testing.T) {
		signer, err := NewSignerFromPrivateKey("invalid-key")
		require.Error(t, err)
		require.Nil(t, signer)
		require.Contains(t, err.Error(), "failed to load private key")
	})

	t.Run("malformed private key format", func(t *testing.T) {
		signer, err := NewSignerFromPrivateKey("PrivateKey-invalid")
		require.Error(t, err)
		require.Nil(t, signer)
		require.Contains(t, err.Error(), "failed to load private key")
	})
}
