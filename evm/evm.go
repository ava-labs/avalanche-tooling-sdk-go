// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/vms/evm/predicate"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/subnet-evm/accounts/abi/bind"
	"github.com/ava-labs/subnet-evm/ethclient"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ava-labs/subnet-evm/plugin/evm/upgrade/legacy"
	"github.com/ava-labs/subnet-evm/precompile/contracts/warp"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	avalancheWarp "github.com/ava-labs/avalanchego/vms/platformvm/warp"
	ethereum "github.com/ava-labs/libevm"
	ethparams "github.com/ava-labs/libevm/params"
)

const (
	repeatsOnFailure            = 3
	baseFeeFactor               = 2
	maxPriorityFeePerGas        = 2500000000 // 2.5 gwei
	nativeTransferGas    uint64 = 21_000
)

// also used at mocks
var (
	ethclientDialContext = ethclient.DialContext
	sleepBetweenRepeats  = 1 * time.Second
)

// wraps over ethclient for calls used by SDK. featues:
// - finds out url scheme in case it is missing, to connect to ws/wss/http/https
// - repeats to try to recover from failures, generating its own context for each call
// - logs rpc url in case of failure
// - receives addresses and private keys as strings
type Client struct {
	EthClient ethclient.Client
	URL       string
}

// indicates if the given rpc url has schema or not
func HasScheme(rpcURL string) (bool, error) {
	if parsedURL, err := url.Parse(rpcURL); err != nil {
		if !strings.Contains(err.Error(), "first path segment in URL cannot contain colon") {
			return false, err
		}
		return false, nil
	} else {
		return strings.Contains(rpcURL, "://") && parsedURL.Scheme != "", nil
	}
}

// tries to connect an ethclient to a rpc url without scheme,
// by trying out different possible schemes: ws, wss, https, http
func GetClientWithoutScheme(rpcURL string) (ethclient.Client, string, error) {
	if b, err := HasScheme(rpcURL); err != nil {
		return nil, "", err
	} else if b {
		return nil, "", fmt.Errorf("url does have scheme")
	}
	notDeterminedErr := fmt.Errorf("url %s has no scheme and protocol could not be determined", rpcURL)
	// let's start with ws it always give same error for http/https/wss
	scheme := "ws://"
	ctx, cancel := utils.GetAPILargeContext()
	defer cancel()
	client, err := ethclientDialContext(ctx, scheme+rpcURL)
	if err == nil {
		return client, scheme, nil
	} else if !strings.Contains(err.Error(), "websocket: bad handshake") {
		return nil, "", notDeterminedErr
	}
	// wss give specific errors for http/http
	scheme = "wss://"
	client, err = ethclientDialContext(ctx, scheme+rpcURL)
	if err == nil {
		return client, scheme, nil
	} else if !strings.Contains(err.Error(), "websocket: bad handshake") && // may be https
		!strings.Contains(err.Error(), "first record does not look like a TLS handshake") { // may be http
		return nil, "", notDeterminedErr
	}
	// https/http discrimination based on sending a specific query
	scheme = "https://"
	client, err = ethclientDialContext(ctx, scheme+rpcURL)
	if err == nil {
		_, err = client.ChainID(ctx)
		switch {
		case err == nil:
			return client, scheme, nil
		case strings.Contains(err.Error(), "server gave HTTP response to HTTPS client"):
			scheme = "http://"
			client, err = ethclientDialContext(ctx, scheme+rpcURL)
			if err == nil {
				return client, scheme, nil
			}
		}
	}
	return nil, "", notDeterminedErr
}

// connects an evm client to the given [rpcURL]
// supports [repeatsOnFailure] failures
func GetClient(rpcURL string) (Client, error) {
	client := Client{
		URL: rpcURL,
	}
	hasScheme, err := HasScheme(rpcURL)
	if err != nil {
		return client, fmt.Errorf("failure determining the scheme of url %s: %w", rpcURL, err)
	}
	client.EthClient, err = utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (ethclient.Client, error) {
			if hasScheme {
				return ethclientDialContext(ctx, rpcURL)
			} else {
				client, _, err := GetClientWithoutScheme(rpcURL)
				return client, err
			}
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure connecting to %s: %w", rpcURL, err)
	}
	return client, err
}

// closes underlying ethclient connection
func (client Client) Close() {
	client.EthClient.Close()
}

// indicates wether a contract is deployed on [contractAddress]
// supports [repeatsOnFailure] failures
func (client Client) ContractAlreadyDeployed(
	contractAddress string,
) (bool, error) {
	if bs, err := client.GetContractBytecode(contractAddress); err != nil {
		return false, err
	} else {
		return len(bs) != 0, nil
	}
}

// returns the contract bytecode at [contractAddress]
// supports [repeatsOnFailure] failures
func (client Client) GetContractBytecode(
	contractAddressStr string,
) ([]byte, error) {
	contractAddress := common.HexToAddress(contractAddressStr)
	code, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) ([]byte, error) {
			return client.EthClient.CodeAt(ctx, contractAddress, nil)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf(
			"failure obtaining code from %s at address %s: %w",
			client.URL,
			contractAddressStr,
			err,
		)
	}
	return code, err
}

// returns the balance for [signer]
// supports [repeatsOnFailure] failures
func (client Client) GetSignerBalance(
	signer *Signer,
) (*big.Int, error) {
	return client.GetAddressBalance(signer.Address().Hex())
}

// returns the balance for [address]
// supports [repeatsOnFailure] failures
func (client Client) GetAddressBalance(
	addressStr string,
) (*big.Int, error) {
	address := common.HexToAddress(addressStr)
	balance, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*big.Int, error) {
			return client.EthClient.BalanceAt(ctx, address, nil)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure obtaining balance for %s on %s: %w", addressStr, client.URL, err)
	}
	return balance, err
}

// returns the nonce at [address]
// supports [repeatsOnFailure] failures
func (client Client) NonceAt(
	addressStr string,
) (uint64, error) {
	address := common.HexToAddress(addressStr)
	nonce, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (uint64, error) {
			return client.EthClient.NonceAt(ctx, address, nil)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure obtaining nonce for %s on %s: %w", addressStr, client.URL, err)
	}
	return nonce, err
}

// returns the suggested gas tip
// supports [repeatsOnFailure] failures
func (client Client) SuggestGasTipCap() (*big.Int, error) {
	gasTipCap, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*big.Int, error) {
			return client.EthClient.SuggestGasTipCap(ctx)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure obtaining gas tip cap on %s: %w", client.URL, err)
	}
	return gasTipCap, err
}

// returns the estimated base fee
// supports [repeatsOnFailure] failures
func (client Client) EstimateBaseFee() (*big.Int, error) {
	baseFee, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*big.Int, error) {
			return client.EthClient.EstimateBaseFee(ctx)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure estimating base fee on %s: %w", client.URL, err)
	}
	return baseFee, err
}

// Returns gasFeeCap, gasTipCap, and nonce to be used when constructing a transaction
// supports [repeatsOnFailure] failures on each step
func (client Client) CalculateTxParams(
	address string,
) (*big.Int, *big.Int, uint64, error) {
	baseFee, err := client.EstimateBaseFee()
	if err != nil {
		return nil, nil, 0, err
	}
	gasTipCap, err := client.SuggestGasTipCap()
	if err != nil {
		return nil, nil, 0, err
	}
	nonce, err := client.NonceAt(address)
	if err != nil {
		return nil, nil, 0, err
	}
	gasFeeCap := baseFee.Mul(baseFee, big.NewInt(baseFeeFactor))
	// Use the max of gasTipCap and the hardcoded minimum to ensure gasFeeCap is always valid
	tipToUse := gasTipCap
	minTip := big.NewInt(maxPriorityFeePerGas)
	if gasTipCap.Cmp(minTip) < 0 {
		tipToUse = minTip
	}
	gasFeeCap.Add(gasFeeCap, tipToUse)
	return gasFeeCap, gasTipCap, nonce, nil
}

// returns the estimated gas limit
// supports [repeatsOnFailure] failures
func (client Client) EstimateGasLimit(
	msg ethereum.CallMsg,
) (uint64, error) {
	gasLimit, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (uint64, error) {
			return client.EthClient.EstimateGas(ctx, msg)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure estimating gas limit on %s: %w", client.URL, err)
	}
	return gasLimit, err
}

// returns the chain ID
// supports [repeatsOnFailure] failures
func (client Client) GetChainID() (*big.Int, error) {
	chainID, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*big.Int, error) {
			return client.EthClient.ChainID(ctx)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure getting chain id from %s: %w", client.URL, err)
	}
	return chainID, err
}

// returns the chain conf
// supports [repeatsOnFailure] failures
func (client Client) ChainConfig() (*params.ChainConfigWithUpgradesJSON, error) {
	conf, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*params.ChainConfigWithUpgradesJSON, error) {
			return client.EthClient.ChainConfig(ctx)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure getting chain config from %s: %w", client.URL, err)
	}
	return conf, err
}

// sends [tx]
// supports [repeatsOnFailure] failures
func (client Client) SendTransaction(
	tx *types.Transaction,
) error {
	_, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (any, error) {
			return nil, client.EthClient.SendTransaction(ctx, tx)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure sending transaction %#v to %s: %w", tx, client.URL, err)
	}
	return err
}

// waits for [tx]'s receipt to have successful state
// supports [repeatsOnFailure] failures
func (client Client) WaitForTransaction(
	tx *types.Transaction,
) (*types.Receipt, bool, error) {
	steps := int(constants.APIRequestLargeTimeout.Seconds())
	var cumErr error
	for step := 0; step < steps; step++ {
		receipt, err := client.TransactionReceipt(tx.Hash())
		if err == nil {
			var success bool
			if receipt != nil {
				success = receipt.Status == types.ReceiptStatusSuccessful
			}
			return receipt, success, err
		}
		cumErr = errors.Join(cumErr, err)
		time.Sleep(sleepBetweenRepeats)
	}
	return nil, false, fmt.Errorf("timeout of %d seconds while waiting for tx %#v on %s: %w", steps, tx, client.URL, cumErr)
}

// transfers [amount] to [targetAddressStr] using [signer]
// supports [repeatsOnFailure] failures on each step
func (client Client) FundAddress(
	signer *Signer,
	targetAddressStr string,
	amount *big.Int,
) (*types.Receipt, error) {
	sourceAddress := signer.Address()
	gasFeeCap, gasTipCap, nonce, err := client.CalculateTxParams(sourceAddress.Hex())
	if err != nil {
		return nil, err
	}
	targetAddress := common.HexToAddress(targetAddressStr)
	chainID, err := client.GetChainID()
	if err != nil {
		return nil, err
	}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &targetAddress,
		Gas:       nativeTransferGas,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Value:     amount,
	})
	signedTx, err := signer.SignTx(chainID, tx)
	if err != nil {
		return nil, err
	}
	if err := client.SendTransaction(signedTx); err != nil {
		return nil, err
	}
	receipt, b, err := client.WaitForTransaction(signedTx)
	if err != nil {
		return nil, err
	} else if !b {
		return nil, fmt.Errorf("failure funding %s from %s amount %d", targetAddressStr, sourceAddress.Hex(), amount)
	}
	return receipt, nil
}

// encode [txStr] to binary, sends and waits for it
// supports [repeatsOnFailure] failures on each step
func (client Client) IssueTx(
	txStr string,
) error {
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(common.FromHex(txStr)); err != nil {
		return err
	}
	if err := client.SendTransaction(tx); err != nil {
		return err
	}
	if receipt, b, err := client.WaitForTransaction(tx); err != nil {
		return err
	} else if !b {
		return fmt.Errorf("failure sending tx: got status %d expected %d", receipt.Status, types.ReceiptStatusSuccessful)
	}
	return nil
}

// returns tx options that include signer for [signer]
// supports [repeatsOnFailure] failures when gathering chain info
func (client Client) GetTxOptsWithSigner(
	signer *Signer,
) (*bind.TransactOpts, error) {
	chainID, err := client.GetChainID()
	if err != nil {
		return nil, fmt.Errorf("failure generating signer: %w", err)
	}

	return signer.TransactOpts(chainID)
}

// waits for [timeout] until evm is bootstrapped
// considers evm is bootstrapped if it responds to an evm call (ChainID)
func (client Client) WaitForEVMBootstrapped(timeout time.Duration) error {
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	steps := int(timeout.Seconds())
	var cumErr error
	for step := 0; step < steps; step++ {
		if _, err := client.GetChainID(); err == nil {
			return nil
		} else {
			cumErr = errors.Join(cumErr, err)
		}
		time.Sleep(sleepBetweenRepeats)
	}
	return fmt.Errorf("client at %s not bootstrapped after %.2f seconds: %w", client.URL, timeout.Seconds(), cumErr)
}

// generates a transaction signed with [signer], calling a [contract] method using [callData]
// including [warpMessage] in the tx accesslist
func (client Client) TransactWithWarpMessage(
	signer *Signer,
	warpMessage *avalancheWarp.Message,
	contract common.Address,
	callData []byte,
	value *big.Int,
) (*types.Transaction, error) {
	from := signer.Address()
	gasFeeCap, gasTipCap, nonce, err := client.CalculateTxParams(from.Hex())
	if err != nil {
		return nil, err
	}
	chainID, err := client.GetChainID()
	if err != nil {
		return nil, err
	}
	accessList := types.AccessList{
		types.AccessTuple{
			Address:     warp.ContractAddress,
			StorageKeys: predicate.New(warpMessage.Bytes()),
		},
	}
	msg := ethereum.CallMsg{
		From:       from,
		To:         &contract,
		GasPrice:   nil,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Value:      value,
		Data:       callData,
		AccessList: accessList,
	}
	gasLimit, err := client.EstimateGasLimit(msg)
	if err != nil {
		return nil, err
	}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		To:         &contract,
		Gas:        gasLimit,
		GasFeeCap:  gasFeeCap,
		GasTipCap:  gasTipCap,
		Value:      value,
		Data:       callData,
		AccessList: accessList,
	})
	return signer.SignTx(chainID, tx)
}

// gets block [n]
// supports [repeatsOnFailure] failures
func (client Client) BlockByNumber(n *big.Int) (*types.Block, error) {
	block, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*types.Block, error) {
			return client.EthClient.BlockByNumber(ctx, n)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure retrieving block %d on %s: %w", n, client.URL, err)
	}
	return block, err
}

// get logs as given by [query]
// supports [repeatsOnFailure] failures
func (client Client) FilterLogs(query ethereum.FilterQuery) ([]types.Log, error) {
	logs, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) ([]types.Log, error) {
			return client.EthClient.FilterLogs(ctx, query)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure retrieving logs on %s: %w", client.URL, err)
	}
	return logs, err
}

// get tx receipt for [hash]
// supports [repeatsOnFailure] failures
func (client Client) TransactionReceipt(hash common.Hash) (*types.Receipt, error) {
	receipt, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*types.Receipt, error) {
			return client.EthClient.TransactionReceipt(ctx, hash)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure retrieving receipt for %s on %s: %w", hash, client.URL, err)
	}
	return receipt, err
}

// gets current height
// supports [repeatsOnFailure] failures
func (client Client) BlockNumber() (uint64, error) {
	blockNumber, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (uint64, error) {
			return client.EthClient.BlockNumber(ctx)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure retrieving height (block number) on %s: %w", client.URL, err)
	}
	return blockNumber, err
}

// waits until current height is bigger than the given previous height at [prevBlockNumber]
// supports [repeatsOnFailure] failures on each step
func (client Client) WaitForNewBlock(
	prevBlockNumber uint64,
	totalDuration time.Duration,
) error {
	if totalDuration == 0 {
		totalDuration = 10 * time.Second
	}
	steps := int(totalDuration.Seconds())
	for step := 0; step < steps; step++ {
		blockNumber, err := client.BlockNumber()
		if err != nil {
			return err
		}
		if blockNumber > prevBlockNumber {
			return nil
		}
		time.Sleep(sleepBetweenRepeats)
	}
	return fmt.Errorf("no new block produced on %s in %d seconds", client.URL, steps)
}

// issue dummy txs to create the given number of blocks
func (client Client) CreateDummyBlocks(
	numBlocks int,
	signer *Signer,
) error {
	addr := signer.Address()
	chainID, err := client.GetChainID()
	if err != nil {
		return err
	}
	gasPrice := big.NewInt(legacy.BaseFee)
	blockNumber, err := client.BlockNumber()
	if err != nil {
		return fmt.Errorf("unable to get block number: %w", err)
	}
	nonce, err := client.NonceAt(addr.Hex())
	if err != nil {
		return fmt.Errorf("unable to get nonce: %w", err)
	}
	for i := 0; i < numBlocks; i++ {
		// it may be the case that we hit an outdated node with the rpc, so lets not fully trust the API
		if blockNumberFromAPI, err := client.BlockNumber(); err != nil {
			return fmt.Errorf("client.BlockNumber failure at step %d: %w", i, err)
		} else if blockNumberFromAPI > blockNumber {
			// changes from outside
			blockNumber = blockNumberFromAPI
		}
		if nonceFromAPI, err := client.NonceAt(addr.Hex()); err != nil {
			return fmt.Errorf("client.NonceAt failure at step %d: %w", i, err)
		} else if nonceFromAPI > nonce {
			// changes from outside
			nonce = nonceFromAPI
		}
		// send Big1 to himself
		tx := types.NewTransaction(nonce, addr, common.Big1, ethparams.TxGas, gasPrice, nil)
		triggerTx, err := signer.SignTx(chainID, tx)
		if err != nil {
			return fmt.Errorf("signer.SignTx failure at step %d: %w", i, err)
		}
		if err := client.SendTransaction(triggerTx); err != nil {
			return fmt.Errorf("client.SendTransaction failure at step %d: %w", i, err)
		}
		if err := client.WaitForNewBlock(blockNumber, 0); err != nil {
			return fmt.Errorf("WaitForNewBlock failure at step %d: %w", i, err)
		}
		blockNumber++
		nonce++
		time.Sleep(5 * time.Second)
	}
	return nil
}

// issue transactions on [client] so as to activate Proposer VM Fork
// this should generate a PostForkBlock because its parent block
// (genesis) has a timestamp (0) that is greater than or equal to the fork
// activation time of 0. Therefore, subsequent blocks should be built with
// BuildBlockWithContext.
// the current timestamp should be after the ProposerVM activation time (aka ApricotPhase4).
// supports [repeatsOnFailure] failures on each step
func (client Client) SetupProposerVM(
	signer *Signer,
) error {
	const numBlocks = 2 // Number of blocks needed to activate the proposer VM fork
	_, err := utils.Retry(
		func() (any, error) {
			return nil, client.CreateDummyBlocks(numBlocks, signer)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure issuing tx to activate proposer VM: %w", err)
	}
	return err
}

// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the blockchain.
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
// supports [repeatsOnFailure] failures
func (client Client) CallContract(
	msg ethereum.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	result, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) ([]byte, error) {
			return client.EthClient.CallContract(ctx, msg, blockNumber)
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure calling contract on %s: %w", client.URL, err)
	}
	return result, err
}

// TransactionByHash returns the transaction with the given hash.
// supports [repeatsOnFailure] failures
func (client Client) TransactionByHash(
	txHash common.Hash,
) (*types.Transaction, bool, error) {
	type result struct {
		tx        *types.Transaction
		isPending bool
	}
	res, err := utils.RetryWithContextGen(
		utils.GetAPILargeContext,
		func(ctx context.Context) (*result, error) {
			tx, isPending, err := client.EthClient.TransactionByHash(ctx, txHash)
			if err != nil {
				return nil, err
			}
			return &result{tx: tx, isPending: isPending}, nil
		},
		repeatsOnFailure,
		sleepBetweenRepeats,
	)
	if err != nil {
		err = fmt.Errorf("failure getting transaction %s on %s: %w", txHash.Hex(), client.URL, err)
		return nil, false, err
	}
	return res.tx, res.isPending, nil
}
