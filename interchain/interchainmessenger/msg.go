// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package interchainmessenger

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/evm"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ethereum/go-ethereum/common"
)

func GetNextMessageID(
	rpcURL string,
	messengerAddress common.Address,
	destinationBlockchainID ids.ID,
) (ids.ID, error) {
	out, err := evm.CallToMethod(
		rpcURL,
		messengerAddress,
		"getNextMessageID(bytes32)->(bytes32)",
		destinationBlockchainID,
	)
	if err != nil {
		return ids.Empty, err
	}
	received, b := out[0].([32]byte)
	if !b {
		return ids.Empty, fmt.Errorf("error at getNextMessageID call, expected ids.ID, got %T", out[0])
	}
	return received, nil
}

func MessageReceived(
	rpcURL string,
	messengerAddress common.Address,
	messageID ids.ID,
) (bool, error) {
	out, err := evm.CallToMethod(
		rpcURL,
		messengerAddress,
		"messageReceived(bytes32)->(bool)",
		messageID,
	)
	if err != nil {
		return false, err
	}
	received, b := out[0].(bool)
	if !b {
		return false, fmt.Errorf("error at messageReceived call, expected bool, got %T", out[0])
	}
	return received, nil
}

func WaitForMessageReception(
	rpcEndpoint string,
	messengerAddress string,
	messageID ids.ID,
	checkInterval time.Duration,
	checkTimeout time.Duration,
) error {
	if checkInterval == 0 {
		checkInterval = 100 * time.Millisecond
	}
	if checkTimeout == 0 {
		checkTimeout = 10 * time.Second
	}
	t0 := time.Now()
	for {
		if b, err := MessageReceived(
			rpcEndpoint,
			common.HexToAddress(messengerAddress),
			messageID,
		); err != nil {
			return err
		} else if b {
			break
		}
		elapsed := time.Since(t0)
		if elapsed > checkTimeout {
			return fmt.Errorf("timeout waiting for icm message id %s to be received", messageID.String())
		}
		time.Sleep(checkInterval)
	}
	return nil
}

func SendCrossChainMessage(
	rpcURL string,
	messengerAddress common.Address,
	privateKey string,
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
	return evm.TxToMethod(
		rpcURL,
		privateKey,
		messengerAddress,
		nil,
		"sendCrossChainMessage((bytes32, address, (address, uint256), uint256, [address], bytes))->(bytes32)",
		params,
	)
}

// events

type TeleporterMessageReceipt struct {
	ReceivedMessageNonce *big.Int
	RelayerRewardAddress common.Address
}
type TeleporterFeeInfo struct {
	FeeTokenAddress common.Address
	Amount          *big.Int
}
type TeleporterMessage struct {
	MessageNonce            *big.Int
	OriginSenderAddress     common.Address
	DestinationBlockchainID [32]byte
	DestinationAddress      common.Address
	RequiredGasLimit        *big.Int
	AllowedRelayerAddresses []common.Address
	Receipts                []TeleporterMessageReceipt
	Message                 []byte
}
type TeleporterMessengerSendCrossChainMessage struct {
	MessageID               [32]byte
	DestinationBlockchainID [32]byte
	Message                 TeleporterMessage
	FeeInfo                 TeleporterFeeInfo
}

func ParseSendCrossChainMessage(log types.Log) (*TeleporterMessengerSendCrossChainMessage, error) {
	event := new(TeleporterMessengerSendCrossChainMessage)
	if err := evm.UnpackLog(
		"SendCrossChainMessage(bytes32,bytes32,(uint256,address,bytes32,address,uint256,[address],[(uint256,address)],bytes),(address,uint256))",
		[]int{0, 1},
		log,
		event,
	); err != nil {
		return nil, err
	}
	return event, nil
}
