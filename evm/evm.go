// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/subnet-evm/accounts/abi/bind"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/ethclient"
	"github.com/ava-labs/subnet-evm/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	BaseFeeFactor               = 2
	MaxPriorityFeePerGas        = 2500000000 // 2.5 gwei
	NativeTransferGas    uint64 = 21_000
	repeatsOnFailure            = 3
)

func ContractAlreadyDeployed(
	client ethclient.Client,
	contractAddress string,
) (bool, error) {
	if bs, err := GetContractBytecode(client, contractAddress); err != nil {
		return false, err
	} else {
		return len(bs) != 0, nil
	}
}

func GetContractBytecode(
	client ethclient.Client,
	contractAddressStr string,
) ([]byte, error) {
	contractAddress := common.HexToAddress(contractAddressStr)
	return utils.Retry(
		func(ctx context.Context) ([]byte, error) { return client.CodeAt(ctx, contractAddress, nil) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure obtaining code for %s on %#v", contractAddressStr, client),
	)
}

func GetAddressBalance(
	client ethclient.Client,
	addressStr string,
) (*big.Int, error) {
	address := common.HexToAddress(addressStr)
	return utils.Retry(
		func(ctx context.Context) (*big.Int, error) { return client.BalanceAt(ctx, address, nil) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure obtaining balance for %s on %#v", addressStr, client),
	)
}

// Returns the gasFeeCap, gasTipCap, and nonce the be used when constructing a transaction from address
func CalculateTxParams(
	client ethclient.Client,
	addressStr string,
) (*big.Int, *big.Int, uint64, error) {
	baseFee, err := EstimateBaseFee(client)
	if err != nil {
		return nil, nil, 0, err
	}
	gasTipCap, err := SuggestGasTipCap(client)
	if err != nil {
		return nil, nil, 0, err
	}
	nonce, err := NonceAt(client, addressStr)
	if err != nil {
		return nil, nil, 0, err
	}
	gasFeeCap := baseFee.Mul(baseFee, big.NewInt(BaseFeeFactor))
	gasFeeCap.Add(gasFeeCap, big.NewInt(MaxPriorityFeePerGas))
	return gasFeeCap, gasTipCap, nonce, nil
}

func NonceAt(
	client ethclient.Client,
	addressStr string,
) (uint64, error) {
	address := common.HexToAddress(addressStr)
	return utils.Retry(
		func(ctx context.Context) (uint64, error) { return client.NonceAt(ctx, address, nil) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure obtaining nonce for %s on %#v", addressStr, client),
	)
}

func SuggestGasTipCap(
	client ethclient.Client,
) (*big.Int, error) {
	return utils.Retry(
		func(ctx context.Context) (*big.Int, error) { return client.SuggestGasTipCap(ctx) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure obtaining gas tip cap on %#v", client),
	)
}

func EstimateBaseFee(
	client ethclient.Client,
) (*big.Int, error) {
	return utils.Retry(
		func(ctx context.Context) (*big.Int, error) { return client.EstimateBaseFee(ctx) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure estimating base fee on %#v", client),
	)
}

func SetMinBalance(
	client ethclient.Client,
	privateKey string,
	address string,
	requiredMinBalance *big.Int,
) error {
	balance, err := GetAddressBalance(client, address)
	if err != nil {
		return err
	}
	if balance.Cmp(requiredMinBalance) < 0 {
		toFund := big.NewInt(0).Sub(requiredMinBalance, balance)
		err := Transfer(
			client,
			privateKey,
			address,
			toFund,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func Transfer(
	client ethclient.Client,
	sourceAddressPrivateKeyStr string,
	targetAddressStr string,
	amount *big.Int,
) error {
	sourceAddressPrivateKey, err := crypto.HexToECDSA(sourceAddressPrivateKeyStr)
	if err != nil {
		return err
	}
	sourceAddress := crypto.PubkeyToAddress(sourceAddressPrivateKey.PublicKey)
	gasFeeCap, gasTipCap, nonce, err := CalculateTxParams(client, sourceAddress.Hex())
	if err != nil {
		return err
	}
	targetAddress := common.HexToAddress(targetAddressStr)
	chainID, err := GetChainID(client)
	if err != nil {
		return err
	}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &targetAddress,
		Gas:       NativeTransferGas,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Value:     amount,
	})
	txSigner := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, txSigner, sourceAddressPrivateKey)
	if err != nil {
		return err
	}
	if err := SendTransaction(client, signedTx); err != nil {
		return err
	}
	if _, b, err := WaitForTransaction(client, signedTx); err != nil {
		return err
	} else if !b {
		return fmt.Errorf("failure funding %s from %s amount %d", targetAddressStr, sourceAddress.Hex(), amount)
	}
	return nil
}

func IssueTx(
	client ethclient.Client,
	txStr string,
) error {
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(common.FromHex(txStr)); err != nil {
		return err
	}
	if err := SendTransaction(client, tx); err != nil {
		return err
	}
	if receipt, b, err := WaitForTransaction(client, tx); err != nil {
		return err
	} else if !b {
		return fmt.Errorf("failure sending tx: got status %d expected %d", receipt.Status, types.ReceiptStatusSuccessful)
	}
	return nil
}

func SendTransaction(
	client ethclient.Client,
	tx *types.Transaction,
) error {
	_, err := utils.Retry(
		func(ctx context.Context) (interface{}, error) { return nil, client.SendTransaction(ctx, tx) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure sending transaction %#v to %#v", tx, client),
	)
	return err
}

func GetClient(rpcURL string) (ethclient.Client, error) {
	return utils.Retry(
		func(ctx context.Context) (ethclient.Client, error) { return ethclient.DialContext(ctx, rpcURL) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure connecting to %s", rpcURL),
	)
}

func GetChainID(client ethclient.Client) (*big.Int, error) {
	return utils.Retry(
		func(ctx context.Context) (*big.Int, error) { return client.ChainID(ctx) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure getting chain id from client %#v", client),
	)
}

func GetTxOptsWithSigner(
	client ethclient.Client,
	prefundedPrivateKeyStr string,
) (*bind.TransactOpts, error) {
	prefundedPrivateKey, err := crypto.HexToECDSA(prefundedPrivateKeyStr)
	if err != nil {
		return nil, err
	}
	chainID, err := GetChainID(client)
	if err != nil {
		return nil, fmt.Errorf("failure generating signer: %w", err)
	}
	return bind.NewKeyedTransactorWithChainID(prefundedPrivateKey, chainID)
}

func WaitForTransaction(
	client ethclient.Client,
	tx *types.Transaction,
) (*types.Receipt, bool, error) {
	receipt, err := utils.Retry(
		func(ctx context.Context) (*types.Receipt, error) { return bind.WaitMined(ctx, client, tx) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure waiting for tx %#v on client %#v", tx, client),
	)
	var success bool
	if receipt != nil {
		success = receipt.Status == types.ReceiptStatusSuccessful
	}
	return receipt, success, err
}

// Returns the first log in 'logs' that is successfully parsed by 'parser'
func GetEventFromLogs[T any](logs []*types.Log, parser func(log types.Log) (T, error)) (T, error) {
	cumErrMsg := ""
	for i, log := range logs {
		event, err := parser(*log)
		if err == nil {
			return event, nil
		}
		if cumErrMsg != "" {
			cumErrMsg += "; "
		}
		cumErrMsg += fmt.Sprintf("log %d -> %s", i, err.Error())
	}
	return *new(T), fmt.Errorf("failed to find %T event in receipt logs: [%s]", *new(T), cumErrMsg)
}

func GetRPCClient(rpcURL string) (*rpc.Client, error) {
	return utils.Retry(
		func(ctx context.Context) (*rpc.Client, error) { return rpc.DialContext(ctx, rpcURL) },
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure connecting to %s", rpcURL),
	)
}

func DebugTraceTransaction(
	client *rpc.Client,
	txID string,
) (map[string]interface{}, error) {
	var trace map[string]interface{}
	_, err := utils.Retry(
		func(ctx context.Context) (interface{}, error) {
			return nil, client.CallContext(
				ctx,
				&trace,
				"debug_traceTransaction",
				txID,
				map[string]string{"tracer": "callTracer"},
			)
		},
		constants.APIRequestLargeTimeout,
		repeatsOnFailure,
		fmt.Sprintf("failure tracing tx %s for client %#v", txID, client),
	)
	return trace, err
}

func GetTrace(rpcURL string, txID string) (map[string]interface{}, error) {
	client, err := GetRPCClient(rpcURL)
	if err != nil {
		return nil, err
	}
	return DebugTraceTransaction(client, txID)
}
