// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package icm

import (
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/libevm/common"
	"github.com/stretchr/testify/require"
)

func TestICMMessageReceipt(t *testing.T) {
	receipt := ICMMessageReceipt{
		ReceivedMessageNonce: big.NewInt(123),
		RelayerRewardAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	require.Equal(t, big.NewInt(123), receipt.ReceivedMessageNonce)
	require.Equal(t, common.HexToAddress("0x1234567890123456789012345678901234567890"), receipt.RelayerRewardAddress)
}

func TestICMFeeInfo(t *testing.T) {
	feeInfo := ICMFeeInfo{
		FeeTokenAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Amount:          big.NewInt(1000),
	}

	require.Equal(t, common.HexToAddress("0x1111111111111111111111111111111111111111"), feeInfo.FeeTokenAddress)
	require.Equal(t, big.NewInt(1000), feeInfo.Amount)
}

func TestICMMessage(t *testing.T) {
	destinationBlockchainID := ids.GenerateTestID()
	message := ICMMessage{
		MessageNonce:            big.NewInt(1),
		OriginSenderAddress:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		DestinationBlockchainID: destinationBlockchainID,
		DestinationAddress:      common.HexToAddress("0x0987654321098765432109876543210987654321"),
		RequiredGasLimit:        big.NewInt(100000),
		AllowedRelayerAddresses: []common.Address{},
		Receipts:                []ICMMessageReceipt{},
		Message:                 []byte("test message"),
	}

	require.Equal(t, big.NewInt(1), message.MessageNonce)
	require.Equal(t, common.HexToAddress("0x1234567890123456789012345678901234567890"), message.OriginSenderAddress)
	require.Equal(t, [32]byte(destinationBlockchainID), message.DestinationBlockchainID)
	require.Equal(t, common.HexToAddress("0x0987654321098765432109876543210987654321"), message.DestinationAddress)
	require.Equal(t, big.NewInt(100000), message.RequiredGasLimit)
	require.Empty(t, message.AllowedRelayerAddresses)
	require.Empty(t, message.Receipts)
	require.Equal(t, []byte("test message"), message.Message)
}

func TestICMMessengerSendCrossChainMessage(t *testing.T) {
	messageID := ids.GenerateTestID()
	destinationBlockchainID := ids.GenerateTestID()

	event := ICMMessengerSendCrossChainMessage{
		MessageID:               messageID,
		DestinationBlockchainID: destinationBlockchainID,
		Message: ICMMessage{
			MessageNonce:            big.NewInt(1),
			OriginSenderAddress:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
			DestinationBlockchainID: destinationBlockchainID,
			DestinationAddress:      common.HexToAddress("0x0987654321098765432109876543210987654321"),
			RequiredGasLimit:        big.NewInt(100000),
			AllowedRelayerAddresses: []common.Address{},
			Receipts:                []ICMMessageReceipt{},
			Message:                 []byte("test message"),
		},
		FeeInfo: ICMFeeInfo{
			FeeTokenAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"),
			Amount:          big.NewInt(1000),
		},
	}

	require.Equal(t, [32]byte(messageID), event.MessageID)
	require.Equal(t, [32]byte(destinationBlockchainID), event.DestinationBlockchainID)
	require.NotNil(t, event.Message)
	require.NotNil(t, event.FeeInfo)
}
