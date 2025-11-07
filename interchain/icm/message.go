// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package icm

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/evm/contract"
)

// GetNextMessageID queries the next message ID that will be assigned to a message
// sent from the messenger contract to the specified destination blockchain.
func GetNextMessageID(
	rpcURL string,
	messengerAddress common.Address,
	destinationBlockchainID ids.ID,
) (ids.ID, error) {
	out, err := contract.CallToMethod(
		rpcURL,
		common.Address{},
		messengerAddress,
		"getNextMessageID(bytes32)->(bytes32)",
		nil,
		destinationBlockchainID,
	)
	if err != nil {
		return ids.Empty, err
	}
	return contract.GetSmartContractCallResult[[32]byte]("getNextMessageID", out)
}

// MessageReceived checks whether a message with the given ID has been received
// and executed by the messenger contract.
func MessageReceived(
	rpcURL string,
	messengerAddress common.Address,
	messageID ids.ID,
) (bool, error) {
	out, err := contract.CallToMethod(
		rpcURL,
		common.Address{},
		messengerAddress,
		"messageReceived(bytes32)->(bool)",
		nil,
		messageID,
	)
	if err != nil {
		return false, err
	}
	return contract.GetSmartContractCallResult[bool]("messageReceived", out)
}

// SendCrossChainMessage sends a cross-chain message from the source chain to a destination chain.
// The message will be delivered to the destination address on the destination blockchain.
// Returns the transaction, receipt, and any error encountered.
func SendCrossChainMessage(
	logger logging.Logger,
	rpcURL string,
	messengerAddress common.Address,
	signer *evm.Signer,
	destinationBlockchainID ids.ID,
	destinationAddress common.Address,
	message []byte,
) (*types.Transaction, *types.Receipt, error) {
	type FeeInfo struct {
		FeeTokenAddress common.Address
		Amount          *big.Int
	}
	type Params struct {
		DestinationBlockchainID [32]byte
		DestinationAddress      common.Address
		FeeInfo                 FeeInfo
		RequiredGasLimit        *big.Int
		AllowedRelayerAddresses []common.Address
		Message                 []byte
	}
	params := Params{
		DestinationBlockchainID: destinationBlockchainID,
		DestinationAddress:      destinationAddress,
		FeeInfo: FeeInfo{
			FeeTokenAddress: common.Address{},
			Amount:          big.NewInt(0),
		},
		RequiredGasLimit:        big.NewInt(1),
		AllowedRelayerAddresses: []common.Address{},
		Message:                 message,
	}
	return contract.TxToMethod(
		logger,
		rpcURL,
		signer,
		messengerAddress,
		nil,
		"send cross chain message",
		nil,
		"sendCrossChainMessage((bytes32, address, (address, uint256), uint256, [address], bytes))->(bytes32)",
		params,
	)
}

// ICMMessageReceipt represents a receipt for a delivered cross-chain message.
type ICMMessageReceipt struct {
	ReceivedMessageNonce *big.Int
	RelayerRewardAddress common.Address
}

// ICMFeeInfo contains fee information for relayer incentivization.
type ICMFeeInfo struct {
	FeeTokenAddress common.Address
	Amount          *big.Int
}

// ICMMessage represents a cross-chain message with all its metadata.
type ICMMessage struct {
	MessageNonce            *big.Int
	OriginSenderAddress     common.Address
	DestinationBlockchainID [32]byte
	DestinationAddress      common.Address
	RequiredGasLimit        *big.Int
	AllowedRelayerAddresses []common.Address
	Receipts                []ICMMessageReceipt
	Message                 []byte
}

// ICMMessengerSendCrossChainMessage represents the SendCrossChainMessage event emitted by TeleporterMessenger.
type ICMMessengerSendCrossChainMessage struct {
	MessageID               [32]byte
	DestinationBlockchainID [32]byte
	Message                 ICMMessage
	FeeInfo                 ICMFeeInfo
}

// ParseSendCrossChainMessage parses a SendCrossChainMessage event from a transaction log.
func ParseSendCrossChainMessage(log types.Log) (*ICMMessengerSendCrossChainMessage, error) {
	event := new(ICMMessengerSendCrossChainMessage)
	if err := contract.UnpackLog(
		"SendCrossChainMessage(bytes32,bytes32,(uint256,address,bytes32,address,uint256,[address],[(uint256,address)],bytes),(address,uint256))",
		[]int{0, 1},
		log,
		event,
	); err != nil {
		return nil, err
	}
	return event, nil
}
