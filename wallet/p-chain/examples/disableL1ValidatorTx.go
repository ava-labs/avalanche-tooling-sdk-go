// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet/p-chain/txs"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
)

func DisableL1Validator() error {
	// Configuration - Replace these with your actual values
	const (
		// Your private key file path
		privateKeyFilePath = "/Users/raymondsukanto/.avalanche-cli/key/newTestKey.pk"

		// Validation ID
		validationIDStr = "2bj5aLW8mCuuEjhGxERaoQvQwM1A1wZBkDeG23JC5MeDLNczQd"
	)

	//// Subnet auth keys (addresses that can sign the conversion tx)
	//subnetAuthKeysStrs := []string{
	//	"P-fujixxx", // Replace with actual addresses
	//}

	network := network.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, privateKeyFilePath, nil)
	if err != nil {
		return fmt.Errorf("failed to create keychain: %w", err)
	}

	// Parse IDs
	validationID, err := ids.FromString(validationIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse subnet ID: %w", err)
	}

	//wallet, err := wallet.New(
	//	context.Background(),
	//	network.Endpoint,
	//	keychain.Keychain,
	//	primary.WalletConfig{
	//		SubnetIDs: []ids.ID{subnetID},
	//	},
	//)
	wallet, err := wallet.New(
		context.Background(),
		network.Endpoint,
		keychain.Keychain,
		primary.WalletConfig{
			ValidationIDs: []ids.ID{validationID},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	disableL1ValidatorTxParams := txs.DisableL1ValidatorTxParams{
		ValidationID: validationID,
		Wallet:       &wallet, // Use the wallet wrapper
	}

	tx, err := txs.NewDisableL1ValidatorTx(disableL1ValidatorTxParams)
	if err != nil {
		return fmt.Errorf("failed to create convert subnet tx: %w", err)
	}

	// Since it has the required signatures, we will now commit the transaction on chain
	txID, err := wallet.Commit(*tx, true)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	fmt.Printf("Disable  L1 transaction submitted successfully! TX ID: %s\n", txID.String())
	return nil
}

func main() {
	if err := DisableL1Validator(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
