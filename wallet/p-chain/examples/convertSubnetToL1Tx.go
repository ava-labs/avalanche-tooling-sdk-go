// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/blockchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/message"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ethereum/go-ethereum/common"
)

func ConvertL1() error {
	// Configuration - Replace these with your actual values
	const (
		// Your private key file path
		privateKeyFilePath = ""

		// Subnet and Chain IDs
		subnetIDStr = ""
		chainIDStr  = ""

		// Validator information
		nodeIDStr            = ""
		BLSPublicKey         = "0x..."     // Replace with actual BLS public key
		BLSProofOfPossession = "0x..."     // Replace with actual BLS proof
		ChangeOwnerAddr      = "P-fujixxx" // Address to receive remaining balance
		Weight               = 100         // Validator weight
		Balance              = 1000000000  // Validator balance in nAVAX

		// Validator manager contract address
		validatorManagerAddr = "0x0FEEDC0DE0000000000000000000000000000000" // Replace with actual contract address
	)

	// Subnet auth keys (addresses that can sign the conversion tx)
	subnetAuthKeysStrs := []string{
		"P-fujixxx", // Replace with actual addresses
	}

	network := network.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, privateKeyFilePath, nil)
	if err != nil {
		return fmt.Errorf("failed to create keychain: %w", err)
	}

	// Parse IDs
	subnetID, err := ids.FromString(subnetIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse subnet ID: %w", err)
	}

	chainID, err := ids.FromString(chainIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse chain ID: %w", err)
	}

	wallet, err := wallet.New(
		context.Background(),
		network.Endpoint,
		keychain.Keychain,
		primary.WalletConfig{
			SubnetIDs: []ids.ID{subnetID},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	subnetAuthKeys, err := address.ParseToIDs(subnetAuthKeysStrs)
	if err != nil {
		return fmt.Errorf("failure parsing auth keys: %w", err)
	}

	bootstrapValidators := []*avagoTxs.ConvertSubnetToL1Validator{}
	nodeID, err := ids.NodeIDFromString(nodeIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse node ID: %w", err)
	}

	blsInfo, err := blockchain.ConvertToBLSProofOfPossession(BLSPublicKey, BLSProofOfPossession)
	if err != nil {
		return fmt.Errorf("failure parsing BLS info: %w", err)
	}

	addrs, err := address.ParseToIDs([]string{ChangeOwnerAddr})
	if err != nil {
		return fmt.Errorf("failure parsing change owner address: %w", err)
	}

	bootstrapValidator := &avagoTxs.ConvertSubnetToL1Validator{
		NodeID:  nodeID[:],
		Weight:  Weight,
		Balance: Balance,
		Signer:  blsInfo,
		RemainingBalanceOwner: message.PChainOwner{
			Threshold: 1,
			Addresses: addrs,
		},
	}
	bootstrapValidators = append(bootstrapValidators, bootstrapValidator)

	convertSubnetParams := txs.ConvertSubnetToL1TxParams{
		// SubnetAuthKeys are the keys used to sign `ConvertSubnetToL1Tx`
		SubnetAuthKeys: subnetAuthKeys,
		// SubnetID is Subnet ID of the subnet to convert to an L1.
		SubnetID: subnetID,
		// ChainID is Blockchain ID of the L1 where the validator manager contract is deployed.
		ChainID: chainID,
		// Address is address of the validator manager contract.
		Address: common.HexToAddress(validatorManagerAddr).Bytes(),
		// Validators are the initial set of L1 validators after the conversion.
		Validators: bootstrapValidators,
		// Wallet is the wallet used to sign `ConvertSubnetToL1Tx`
		Wallet: &wallet, // Use the wallet wrapper
	}

	tx, err := txs.NewConvertSubnetToL1Tx(convertSubnetParams)
	if err != nil {
		return fmt.Errorf("failed to create convert subnet tx: %w", err)
	}

	// Since it has the required signatures, we will now commit the transaction on chain
	txID, err := wallet.Commit(*tx, true)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	fmt.Printf("Convert subnet to L1 transaction submitted successfully! TX ID: %s\n", txID.String())
	return nil
}

func main() {
	if err := ConvertL1(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
