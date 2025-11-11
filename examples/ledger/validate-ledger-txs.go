//go:build validate_ledger_txs
// +build validate_ledger_txs

// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// validate-ledger-txs is a comprehensive example demonstrating how to use the Ledger keychain
// to build and sign all supported Avalanche transaction types across P-Chain, X-Chain, and C-Chain.
//
// This example covers:
//   - P-Chain: Subnet operations (CreateSubnet, ConvertSubnetToL1), validator management
//     (AddValidator, AddSubnetValidator, RemoveSubnetValidator), L1 validator operations
//     (RegisterL1Validator, SetL1ValidatorWeight, IncreaseL1ValidatorBalance, DisableL1Validator),
//     cross-chain transfers (ExportTx, ImportTx), and ownership transfers (TransferSubnetOwnership)
//   - X-Chain: Asset operations (BaseTx, CreateAssetTx, OperationTx) and cross-chain transfers
//     (ImportTx, ExportTx)
//   - C-Chain: Atomic transactions for cross-chain transfers (ImportTx, ExportTx)
//
// The example connects to both a local Avalanche network and Fuji testnet. The local network is
// used for P-Chain operations, while Fuji testnet is used for X-Chain and C-Chain operations
// (as custom local network operations for X-Chain and C-Chain are not yet implemented).
// Some transactions are issued to the network while others are only built and signed for validation purposes.
//
// Prerequisites:
// - Ledger device connected and unlocked
// - Avalanche app open on the Ledger
// - Local Avalanche network running
// - Sufficient AVAX balance on local network for the Ledger address
// - Sufficient AVAX balance on Fuji testnet for the Ledger address
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto/bls/signer/localsigner"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/message"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/payload"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"github.com/ava-labs/coreth/plugin/evm/atomic"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger"

	avmtxs "github.com/ava-labs/avalanchego/vms/avm/txs"
	ethcommon "github.com/ava-labs/libevm/common"
)

const (
	localEndpoint = "http://localhost:9650"
	fujiEndpoint  = "https://api.avax-test.network"
)

func main() {
	fmt.Println("Setting up Ledger connection...")
	fmt.Println("Make sure your Ledger device is connected and the Avalanche app is open.")

	ledgerDevice, err := ledger.New()
	if err != nil {
		log.Fatalf("Failed to connect to Ledger device: %v", err)
	}
	defer ledgerDevice.Disconnect()

	fmt.Println("Ledger device connected successfully!")

	kc, err := ledger.NewKeychainFromIndices(ledgerDevice, []uint32{0})
	if err != nil {
		log.Fatalf("Failed to create ledger keychain: %v", err)
	}

	fmt.Println("\nCreating wallet context...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wallet, err := primary.MakeWallet(
		ctx,
		localEndpoint,
		kc,
		kc,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	fmt.Println("Wallet created successfully!")

	fmt.Println("\nCreating Fuji wallet context...")
	fujiCtx, fujiCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer fujiCancel()

	fujiWallet, err := primary.MakeWallet(
		fujiCtx,
		fujiEndpoint,
		kc,
		kc,
		primary.WalletConfig{},
	)
	if err != nil {
		log.Fatalf("Failed to create fuji wallet: %v", err)
	}
	fmt.Println("Fuji wallet created successfully!")

	controlKeys := kc.Addresses().List()
	threshold := uint32(1)

	owners := &secp256k1fx.OutputOwners{
		Addrs:     controlKeys,
		Threshold: threshold,
		Locktime:  0,
	}

	fmt.Println("\n=== P-Chain Transaction Examples ===")

	// 1. CreateSubnet - Build, Sign and Issue
	fmt.Println("\n1. Building and issuing CreateSubnet transaction...")
	createSubnetTx, err := buildAndSignCreateSubnetTx(wallet, owners)
	if err != nil {
		log.Fatalf("Failed to create CreateSubnet tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", createSubnetTx.ID())

	fmt.Println("   Issuing CreateSubnet transaction to network...")
	issueCtx, issueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer issueCancel()

	if err := wallet.P().IssueTx(createSubnetTx, common.WithContext(issueCtx)); err != nil {
		log.Fatalf("Failed to issue CreateSubnet tx: %v", err)
	}
	fmt.Println("   ✓ CreateSubnet transaction issued successfully!")

	// Use the CreateSubnet tx ID as the subnet ID for subsequent operations
	subnetID := createSubnetTx.ID()
	fmt.Printf("   Subnet ID: %s\n", subnetID)

	// 2. CreateChain
	fmt.Println("\n2. Building CreateChain transaction...")
	createChainTx, err := buildAndSignCreateChainTx(wallet, subnetID)
	if err != nil {
		log.Fatalf("Failed to create CreateChain tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", createChainTx.ID())

	// 3. AddSubnetValidator
	fmt.Println("\n3. Building AddSubnetValidator transaction...")
	addValidatorTx, err := buildAndSignAddSubnetValidatorTx(wallet, subnetID, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create AddSubnetValidator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", addValidatorTx.ID())

	// 4. RemoveSubnetValidator
	fmt.Println("\n4. Building RemoveSubnetValidator transaction...")
	removeValidatorTx, err := buildAndSignRemoveSubnetValidatorTx(wallet, subnetID)
	if err != nil {
		log.Fatalf("Failed to create RemoveSubnetValidator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", removeValidatorTx.ID())

	// 5. TransferSubnetOwnership
	fmt.Println("\n5. Building TransferSubnetOwnership transaction...")
	transferOwnershipTx, err := buildAndSignTransferSubnetOwnershipTx(wallet, subnetID, owners)
	if err != nil {
		log.Fatalf("Failed to create TransferSubnetOwnership tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", transferOwnershipTx.ID())

	// 6. ConvertSubnetToL1 - Build, Sign and Issue
	fmt.Println("\n6. Building and issuing ConvertSubnetToL1 transaction...")
	convertL1Tx, err := buildAndSignConvertSubnetToL1Tx(wallet, subnetID, controlKeys)
	if err != nil {
		log.Fatalf("Failed to create ConvertSubnetToL1 tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", convertL1Tx.ID())

	fmt.Println("   Issuing ConvertSubnetToL1 transaction to network...")
	convertIssueCtx, convertIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer convertIssueCancel()

	if err := wallet.P().IssueTx(convertL1Tx, common.WithContext(convertIssueCtx)); err != nil {
		log.Fatalf("Failed to issue ConvertSubnetToL1 tx: %v", err)
	}
	fmt.Println("   ✓ ConvertSubnetToL1 transaction issued successfully!")

	// Generate validation ID for the first bootstrap validator (index 0)
	// This follows the pattern: validationID = subnetID.Append(index)
	validationID := subnetID.Append(0)
	fmt.Printf("   Bootstrap Validation ID (index 0): %s\n", validationID)

	// 7. BaseTx
	fmt.Println("\n7. Building BaseTx transaction...")
	baseTx, err := buildAndSignBaseTx(wallet, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create BaseTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", baseTx.ID())

	// 8. AddValidator (Primary Network)
	fmt.Println("\n8. Building AddValidator transaction...")
	addPrimaryValidatorTx, err := buildAndSignAddValidatorTx(wallet, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create AddValidator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", addPrimaryValidatorTx.ID())

	// 9. AddDelegator
	fmt.Println("\n9. Building AddDelegator transaction...")
	addDelegatorTx, err := buildAndSignAddDelegatorTx(wallet, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create AddDelegator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", addDelegatorTx.ID())

	// 10. ExportTx (P-Chain to X-Chain) - Using Fuji network for known chain IDs
	fmt.Println("\n10. Building and issuing ExportTx transaction (using Fuji network)...")
	exportTx, err := buildAndSignPChainExportTx(fujiWallet, owners, fujiWallet.X().Builder().Context().BlockchainID)
	if err != nil {
		log.Fatalf("Failed to create ExportTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", exportTx.ID())

	fmt.Println("   Issuing P-Chain ExportTx to network...")
	pExportIssueCtx, pExportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer pExportIssueCancel()

	if err := fujiWallet.P().IssueTx(exportTx, common.WithContext(pExportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue P-Chain ExportTx: %v", err)
	}
	fmt.Println("   ✓ P-Chain ExportTx issued successfully!")

	// 11. X-Chain Import (from P-Chain)
	fmt.Println("\n11. Building and issuing X-Chain ImportTx (from P-Chain)...")
	xChainImportTx, err := buildAndSignXChainImportTx(fujiWallet, owners, constants.PlatformChainID)
	if err != nil {
		log.Fatalf("Failed to create X-Chain ImportTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", xChainImportTx.ID())

	fmt.Println("   Issuing X-Chain ImportTx to network...")
	xImportIssueCtx, xImportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer xImportIssueCancel()

	if err := fujiWallet.X().IssueTx(xChainImportTx, common.WithContext(xImportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue X-Chain ImportTx: %v", err)
	}
	fmt.Println("   ✓ X-Chain ImportTx issued successfully!")

	// 12. X-Chain Export (required for ImportTx)
	fmt.Println("\n12. Building and issuing X-Chain ExportTx (to enable P-Chain import)...")
	xChainExportTx, err := buildAndSignXChainExportTx(fujiWallet, owners, constants.PlatformChainID)
	if err != nil {
		log.Fatalf("Failed to create X-Chain ExportTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", xChainExportTx.ID())

	fmt.Println("   Issuing X-Chain ExportTx to network...")
	xExportIssueCtx, xExportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer xExportIssueCancel()

	if err := fujiWallet.X().IssueTx(xChainExportTx, common.WithContext(xExportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue X-Chain ExportTx: %v", err)
	}
	fmt.Println("   ✓ X-Chain ExportTx issued successfully!")

	// 13. X-Chain BaseTx
	fmt.Println("\n13. Building X-Chain BaseTx (without issuance)...")
	xChainBaseTx, err := buildAndSignXChainBaseTx(fujiWallet, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create X-Chain BaseTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", xChainBaseTx.ID())

	// 14. X-Chain CreateAssetTx
	fmt.Println("\n14. Building X-Chain CreateAssetTx (without issuance)...")
	xChainCreateAssetTx, err := buildAndSignXChainCreateAssetTx(fujiWallet, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create X-Chain CreateAssetTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", xChainCreateAssetTx.ID())

	// 15. X-Chain OperationTx
	fmt.Println("\n15. Building X-Chain OperationTx (without issuance)...")
	xChainOperationTx, err := buildAndSignXChainOperationTx(fujiWallet)
	if err != nil {
		log.Fatalf("Failed to create X-Chain OperationTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", xChainOperationTx.ID())

	// 16. P-Chain ExportTx to C-Chain
	fmt.Println("\n16. Building and issuing P-Chain ExportTx to C-Chain...")
	cChainID := fujiWallet.C().Builder().Context().BlockchainID
	pToC_ExportTx, err := buildAndSignPChainExportTx(fujiWallet, owners, cChainID)
	if err != nil {
		log.Fatalf("Failed to create P-Chain ExportTx to C-Chain: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", pToC_ExportTx.ID())

	fmt.Println("   Issuing P-Chain ExportTx to C-Chain to network...")
	pToC_ExportIssueCtx, pToC_ExportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer pToC_ExportIssueCancel()

	if err := fujiWallet.P().IssueTx(pToC_ExportTx, common.WithContext(pToC_ExportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue P-Chain ExportTx to C-Chain: %v", err)
	}
	fmt.Println("   ✓ P-Chain ExportTx to C-Chain issued successfully!")

	// 17. C-Chain ImportTx from P-Chain
	fmt.Println("\n17. Building and issuing C-Chain ImportTx from P-Chain...")
	pubKey, err := ledgerDevice.PubKey(0)
	if err != nil {
		log.Fatalf("Failed to get pubKey from Ledger: %v", err)
	}
	ethAddr := pubKey.EthAddress()
	fmt.Println("ETH ADDR:", ethAddr)
	cChainImportTx, err := buildAndSignCChainImportTx(fujiWallet, ethAddr, constants.PlatformChainID)
	if err != nil {
		log.Fatalf("Failed to create C-Chain ImportTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", cChainImportTx.ID())

	fmt.Println("   Issuing C-Chain ImportTx to network...")
	cImportIssueCtx, cImportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cImportIssueCancel()

	if err := fujiWallet.C().IssueAtomicTx(cChainImportTx, common.WithContext(cImportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue C-Chain ImportTx: %v", err)
	}
	fmt.Println("   ✓ C-Chain ImportTx issued successfully!")

	// 18. C-Chain ExportTx to P-Chain
	fmt.Println("\n18. Building and issuing C-Chain ExportTx to P-Chain...")
	cChainExportTx, err := buildAndSignCChainExportTx(fujiWallet, owners, constants.PlatformChainID)
	if err != nil {
		log.Fatalf("Failed to create C-Chain ExportTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", cChainExportTx.ID())

	fmt.Println("   Issuing C-Chain ExportTx to network...")
	cExportIssueCtx, cExportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cExportIssueCancel()

	if err := fujiWallet.C().IssueAtomicTx(cChainExportTx, common.WithContext(cExportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue C-Chain ExportTx: %v", err)
	}
	fmt.Println("   ✓ C-Chain ExportTx issued successfully!")

	// 19. P-Chain ImportTx from C-Chain
	fmt.Println("\n19. Building and issuing P-Chain ImportTx from C-Chain...")
	cToP_ImportTx, err := buildAndSignPChainImportTx(fujiWallet, owners, cChainID)
	if err != nil {
		log.Fatalf("Failed to create P-Chain ImportTx from C-Chain: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", cToP_ImportTx.ID())

	fmt.Println("   Issuing P-Chain ImportTx from C-Chain to network...")
	cToP_ImportIssueCtx, cToP_ImportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cToP_ImportIssueCancel()

	if err := fujiWallet.P().IssueTx(cToP_ImportTx, common.WithContext(cToP_ImportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue P-Chain ImportTx from C-Chain: %v", err)
	}
	fmt.Println("   ✓ P-Chain ImportTx from C-Chain issued successfully!")

	// 20. P-Chain ImportTx from X-Chain
	fmt.Println("\n20. Building and issuing P-Chain ImportTx from X-Chain...")
	importTx, err := buildAndSignPChainImportTx(fujiWallet, owners, fujiWallet.X().Builder().Context().BlockchainID)
	if err != nil {
		log.Fatalf("Failed to create ImportTx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", importTx.ID())

	fmt.Println("   Issuing P-Chain ImportTx to network...")
	pImportIssueCtx, pImportIssueCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer pImportIssueCancel()

	if err := fujiWallet.P().IssueTx(importTx, common.WithContext(pImportIssueCtx)); err != nil {
		log.Fatalf("Failed to issue P-Chain ImportTx: %v", err)
	}
	fmt.Println("   ✓ P-Chain ImportTx issued successfully!")

	// 21. AddPermissionlessValidator
	fmt.Println("\n21. Building AddPermissionlessValidator transaction...")
	addPermissionlessValidatorTx, err := buildAndSignAddPermissionlessValidatorTx(wallet, subnetID, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create AddPermissionlessValidator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", addPermissionlessValidatorTx.ID())

	// 22. AddPermissionlessDelegator
	fmt.Println("\n22. Building AddPermissionlessDelegator transaction...")
	addPermissionlessDelegatorTx, err := buildAndSignAddPermissionlessDelegatorTx(wallet, subnetID, controlKeys[0])
	if err != nil {
		log.Fatalf("Failed to create AddPermissionlessDelegator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", addPermissionlessDelegatorTx.ID())

	// 23. TransformSubnet
	fmt.Println("\n23. Building TransformSubnet transaction...")
	transformSubnetTx, err := buildAndSignTransformSubnetTx(wallet, subnetID)
	if err != nil {
		log.Fatalf("Failed to create TransformSubnet tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", transformSubnetTx.ID())

	// 24. RegisterL1Validator
	fmt.Println("\n24. Building RegisterL1Validator transaction...")
	registerL1ValidatorTx, err := buildAndSignRegisterL1ValidatorTx(wallet)
	if err != nil {
		log.Fatalf("Failed to create RegisterL1Validator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", registerL1ValidatorTx.ID())

	// 25. SetL1ValidatorWeight
	fmt.Println("\n25. Building SetL1ValidatorWeight transaction...")
	setL1ValidatorWeightTx, err := buildAndSignSetL1ValidatorWeightTx(wallet)
	if err != nil {
		log.Fatalf("Failed to create SetL1ValidatorWeight tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", setL1ValidatorWeightTx.ID())

	// 26. IncreaseL1ValidatorBalance
	fmt.Println("\n26. Building IncreaseL1ValidatorBalance transaction...")
	increaseL1ValidatorBalanceTx, err := buildAndSignIncreaseL1ValidatorBalanceTx(wallet, validationID)
	if err != nil {
		log.Fatalf("Failed to create IncreaseL1ValidatorBalance tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", increaseL1ValidatorBalanceTx.ID())

	// 27. DisableL1Validator
	fmt.Println("\n27. Building DisableL1Validator transaction...")
	disableL1ValidatorTx, err := buildAndSignDisableL1ValidatorTx(wallet, validationID)
	if err != nil {
		log.Fatalf("Failed to create DisableL1Validator tx: %v", err)
	}
	fmt.Printf("   Tx ID: %s\n", disableL1ValidatorTx.ID())

	fmt.Println("\n✓ All transactions created and signed successfully!")
	fmt.Println("Note: Some transactions were issued to the network for demonstration. Others were only built and signed.")
}

func buildAndSignCreateSubnetTx(wallet *primary.Wallet, owners *secp256k1fx.OutputOwners) (*txs.Tx, error) {
	unsignedTx, err := wallet.P().Builder().NewCreateSubnetTx(owners)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign CreateSubnet transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignCreateChainTx(wallet *primary.Wallet, subnetID ids.ID) (*txs.Tx, error) {
	genesis := []byte("{\"config\":{\"chainId\":1}}")
	vmID := ids.GenerateTestID()
	fxIDs := []ids.ID{}
	chainName := "example-chain"

	unsignedTx, err := wallet.P().Builder().NewCreateChainTx(
		subnetID,
		genesis,
		vmID,
		fxIDs,
		chainName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign CreateChain transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignAddSubnetValidatorTx(wallet *primary.Wallet, subnetID ids.ID, nodeAddr ids.ShortID) (*txs.Tx, error) {
	nodeID := ids.BuildTestNodeID(nodeAddr[:])
	startTime := time.Now().Add(1 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)
	weight := uint64(1000)

	validator := &txs.SubnetValidator{
		Validator: txs.Validator{
			NodeID: nodeID,
			Start:  uint64(startTime.Unix()),
			End:    uint64(endTime.Unix()),
			Wght:   weight,
		},
		Subnet: subnetID,
	}

	unsignedTx, err := wallet.P().Builder().NewAddSubnetValidatorTx(validator)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign AddSubnetValidator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignRemoveSubnetValidatorTx(wallet *primary.Wallet, subnetID ids.ID) (*txs.Tx, error) {
	nodeID := ids.GenerateTestNodeID()

	unsignedTx, err := wallet.P().Builder().NewRemoveSubnetValidatorTx(nodeID, subnetID)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign RemoveSubnetValidator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignTransferSubnetOwnershipTx(wallet *primary.Wallet, subnetID ids.ID, newOwner *secp256k1fx.OutputOwners) (*txs.Tx, error) {
	unsignedTx, err := wallet.P().Builder().NewTransferSubnetOwnershipTx(subnetID, newOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign TransferSubnetOwnership transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignConvertSubnetToL1Tx(wallet *primary.Wallet, subnetID ids.ID, controlKeys []ids.ShortID) (*txs.Tx, error) {
	validatorManagerChainID := ids.GenerateTestID()
	validatorManagerAddress := []byte{1, 2, 3, 4}

	sk, err := localsigner.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate BLS key: %w", err)
	}

	pop, err := signer.NewProofOfPossession(sk)
	if err != nil {
		return nil, fmt.Errorf("failed to create proof of possession: %w", err)
	}

	validators := []*txs.ConvertSubnetToL1Validator{
		{
			NodeID:  ids.GenerateTestNodeID().Bytes(),
			Weight:  1000,
			Balance: 100000,
			Signer:  *pop,
			RemainingBalanceOwner: message.PChainOwner{
				Threshold: 1,
				Addresses: []ids.ShortID{controlKeys[0]},
			},
			DeactivationOwner: message.PChainOwner{
				Threshold: 1,
				Addresses: []ids.ShortID{controlKeys[0]},
			},
		},
	}

	unsignedTx, err := wallet.P().Builder().NewConvertSubnetToL1Tx(
		subnetID,
		validatorManagerChainID,
		validatorManagerAddress,
		validators,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign ConvertSubnetToL1 transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignBaseTx(wallet *primary.Wallet, toAddr ids.ShortID) (*txs.Tx, error) {
	outputs := []*avax.TransferableOutput{
		{
			Asset: avax.Asset{ID: wallet.P().Builder().Context().AVAXAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: 1000000,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{toAddr},
				},
			},
		},
	}

	unsignedTx, err := wallet.P().Builder().NewBaseTx(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign BaseTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignAddValidatorTx(wallet *primary.Wallet, rewardsAddr ids.ShortID) (*txs.Tx, error) {
	nodeID := ids.GenerateTestNodeID()
	startTime := time.Now().Add(1 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)
	weight := uint64(2000000000)
	shares := uint32(20000)

	validator := &txs.Validator{
		NodeID: nodeID,
		Start:  uint64(startTime.Unix()),
		End:    uint64(endTime.Unix()),
		Wght:   weight,
	}

	rewardsOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{rewardsAddr},
	}

	unsignedTx, err := wallet.P().Builder().NewAddValidatorTx(validator, rewardsOwner, shares)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign AddValidator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignAddDelegatorTx(wallet *primary.Wallet, rewardsAddr ids.ShortID) (*txs.Tx, error) {
	nodeID := ids.GenerateTestNodeID()
	startTime := time.Now().Add(1 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)
	weight := uint64(25000000)

	validator := &txs.Validator{
		NodeID: nodeID,
		Start:  uint64(startTime.Unix()),
		End:    uint64(endTime.Unix()),
		Wght:   weight,
	}

	rewardsOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{rewardsAddr},
	}

	unsignedTx, err := wallet.P().Builder().NewAddDelegatorTx(validator, rewardsOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign AddDelegator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignXChainImportTx(wallet *primary.Wallet, to *secp256k1fx.OutputOwners, sourceChainID ids.ID) (*avmtxs.Tx, error) {
	unsignedTx, err := wallet.X().Builder().NewImportTx(
		sourceChainID,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &avmtxs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign X-Chain ImportTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.X().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignXChainExportTx(wallet *primary.Wallet, to *secp256k1fx.OutputOwners, destinationChainID ids.ID) (*avmtxs.Tx, error) {
	unsignedTx, err := wallet.X().Builder().NewExportTx(
		destinationChainID,
		[]*avax.TransferableOutput{
			{
				Asset: avax.Asset{ID: wallet.X().Builder().Context().AVAXAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt:          500000000,
					OutputOwners: *to,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &avmtxs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign X-Chain ExportTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.X().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignXChainBaseTx(wallet *primary.Wallet, toAddr ids.ShortID) (*avmtxs.Tx, error) {
	unsignedTx, err := wallet.X().Builder().NewBaseTx(
		[]*avax.TransferableOutput{
			{
				Asset: avax.Asset{ID: wallet.X().Builder().Context().AVAXAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1000000,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{toAddr},
					},
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &avmtxs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign X-Chain BaseTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.X().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignXChainCreateAssetTx(wallet *primary.Wallet, addr ids.ShortID) (*avmtxs.Tx, error) {
	unsignedTx, err := wallet.X().Builder().NewCreateAssetTx(
		"Test Asset",
		"TST",
		9,
		map[uint32][]verify.State{
			0: {
				&secp256k1fx.TransferOutput{
					Amt: 1000000000,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &avmtxs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign X-Chain CreateAssetTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.X().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignXChainOperationTx(wallet *primary.Wallet) (*avmtxs.Tx, error) {
	unsignedTx, err := wallet.X().Builder().NewOperationTx(
		[]*avmtxs.Operation{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &avmtxs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign X-Chain OperationTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.X().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignPChainImportTx(wallet *primary.Wallet, to *secp256k1fx.OutputOwners, sourceChainID ids.ID) (*txs.Tx, error) {
	unsignedTx, err := wallet.P().Builder().NewImportTx(
		sourceChainID,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign ImportTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignPChainExportTx(wallet *primary.Wallet, to *secp256k1fx.OutputOwners, destinationChainID ids.ID) (*txs.Tx, error) {
	outputs := []*avax.TransferableOutput{
		{
			Asset: avax.Asset{ID: wallet.P().Builder().Context().AVAXAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt:          1000000000,
				OutputOwners: *to,
			},
		},
	}

	unsignedTx, err := wallet.P().Builder().NewExportTx(
		destinationChainID,
		outputs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign ExportTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignAddPermissionlessValidatorTx(wallet *primary.Wallet, subnetID ids.ID, rewardsAddr ids.ShortID) (*txs.Tx, error) {
	nodeID := ids.GenerateTestNodeID()
	startTime := uint64(time.Now().Add(1 * time.Hour).Unix())
	endTime := uint64(time.Now().Add(25 * time.Hour).Unix())
	weight := uint64(1000)

	validator := &txs.SubnetValidator{
		Validator: txs.Validator{
			NodeID: nodeID,
			Start:  startTime,
			End:    endTime,
			Wght:   weight,
		},
		Subnet: subnetID,
	}

	rewardsOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{rewardsAddr},
	}

	unsignedTx, err := wallet.P().Builder().NewAddPermissionlessValidatorTx(
		validator,
		&signer.Empty{},
		wallet.P().Builder().Context().AVAXAssetID,
		rewardsOwner,
		rewardsOwner,
		20000,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign AddPermissionlessValidator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignAddPermissionlessDelegatorTx(wallet *primary.Wallet, subnetID ids.ID, rewardsAddr ids.ShortID) (*txs.Tx, error) {
	nodeID := ids.GenerateTestNodeID()
	startTime := uint64(time.Now().Add(1 * time.Hour).Unix())
	endTime := uint64(time.Now().Add(25 * time.Hour).Unix())
	weight := uint64(500)

	validator := &txs.SubnetValidator{
		Validator: txs.Validator{
			NodeID: nodeID,
			Start:  startTime,
			End:    endTime,
			Wght:   weight,
		},
		Subnet: subnetID,
	}

	rewardsOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{rewardsAddr},
	}

	unsignedTx, err := wallet.P().Builder().NewAddPermissionlessDelegatorTx(
		validator,
		wallet.P().Builder().Context().AVAXAssetID,
		rewardsOwner,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign AddPermissionlessDelegator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignRegisterL1ValidatorTx(wallet *primary.Wallet) (*txs.Tx, error) {
	balance := uint64(100000)
	proofOfPossession := [96]byte{}

	subnetID := ids.GenerateTestID()
	nodeID := ids.GenerateTestNodeID()
	blsPublicKey := [48]byte{}
	expiry := uint64(time.Now().Add(24 * time.Hour).Unix())
	weight := uint64(1000)

	balanceOwners := message.PChainOwner{
		Threshold: 1,
		Addresses: []ids.ShortID{ids.GenerateTestShortID()},
	}
	disableOwners := message.PChainOwner{
		Threshold: 1,
		Addresses: []ids.ShortID{ids.GenerateTestShortID()},
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
	if err != nil {
		return nil, fmt.Errorf("failed to create register L1 validator message: %w", err)
	}

	managerAddress := ids.ShortID{}
	addressedCall, err := payload.NewAddressedCall(
		managerAddress[:],
		addressedCallPayload.Bytes(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create addressed call: %w", err)
	}

	unsignedWarpMessage, err := warp.NewUnsignedMessage(
		wallet.P().Builder().Context().NetworkID,
		ids.GenerateTestID(),
		addressedCall.Bytes(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create unsigned warp message: %w", err)
	}

	emptySignature := &warp.BitSetSignature{
		Signers:   []byte{},
		Signature: [96]byte{},
	}

	signedWarpMessage, err := warp.NewMessage(
		unsignedWarpMessage,
		emptySignature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed warp message: %w", err)
	}

	warpMessage := signedWarpMessage.Bytes()

	unsignedTx, err := wallet.P().Builder().NewRegisterL1ValidatorTx(
		balance,
		proofOfPossession,
		warpMessage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign RegisterL1Validator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignSetL1ValidatorWeightTx(wallet *primary.Wallet) (*txs.Tx, error) {
	validationID := ids.GenerateTestID()
	nonce := uint64(1)
	weight := uint64(2000)

	l1ValidatorWeightPayload, err := message.NewL1ValidatorWeight(
		validationID,
		nonce,
		weight,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 validator weight message: %w", err)
	}

	managerAddress := ids.ShortID{}
	addressedCall, err := payload.NewAddressedCall(
		managerAddress[:],
		l1ValidatorWeightPayload.Bytes(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create addressed call: %w", err)
	}

	unsignedWarpMessage, err := warp.NewUnsignedMessage(
		wallet.P().Builder().Context().NetworkID,
		ids.GenerateTestID(),
		addressedCall.Bytes(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create unsigned warp message: %w", err)
	}

	emptySignature := &warp.BitSetSignature{
		Signers:   []byte{},
		Signature: [96]byte{},
	}

	signedWarpMessage, err := warp.NewMessage(
		unsignedWarpMessage,
		emptySignature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed warp message: %w", err)
	}

	warpMessage := signedWarpMessage.Bytes()

	unsignedTx, err := wallet.P().Builder().NewSetL1ValidatorWeightTx(warpMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign SetL1ValidatorWeight transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignIncreaseL1ValidatorBalanceTx(wallet *primary.Wallet, validationID ids.ID) (*txs.Tx, error) {
	balance := uint64(50000)

	unsignedTx, err := wallet.P().Builder().NewIncreaseL1ValidatorBalanceTx(validationID, balance)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign IncreaseL1ValidatorBalance transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignDisableL1ValidatorTx(wallet *primary.Wallet, validationID ids.ID) (*txs.Tx, error) {
	unsignedTx, err := wallet.P().Builder().NewDisableL1ValidatorTx(validationID)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign DisableL1Validator transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignTransformSubnetTx(wallet *primary.Wallet, subnetID ids.ID) (*txs.Tx, error) {
	assetID := ids.Empty
	initialSupply := uint64(100)
	maxSupply := uint64(100)
	minConsumptionRate := uint64(0)
	maxConsumptionRate := uint64(0)
	minValidatorStake := uint64(1)
	maxValidatorStake := uint64(1000000)
	minStakeDuration := 24 * time.Hour
	maxStakeDuration := 365 * 24 * time.Hour
	minDelegationFee := uint32(20000)
	minDelegatorStake := uint64(1)
	maxValidatorWeightFactor := byte(5)
	uptimeRequirement := uint32(800000)

	unsignedTx, err := wallet.P().Builder().NewTransformSubnetTx(
		subnetID,
		assetID,
		initialSupply,
		maxSupply,
		minConsumptionRate,
		maxConsumptionRate,
		minValidatorStake,
		maxValidatorStake,
		minStakeDuration,
		maxStakeDuration,
		minDelegationFee,
		minDelegatorStake,
		maxValidatorWeightFactor,
		uptimeRequirement,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &txs.Tx{Unsigned: unsignedTx}

	fmt.Println("   *** Please sign TransformSubnet transaction on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.P().Signer().Sign(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignCChainImportTx(wallet *primary.Wallet, to ethcommon.Address, sourceChainID ids.ID) (*atomic.Tx, error) {
	baseFee := big.NewInt(25000000000)
	unsignedTx, err := wallet.C().Builder().NewImportTx(
		sourceChainID,
		to,
		baseFee,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &atomic.Tx{UnsignedAtomicTx: unsignedTx}

	fmt.Println("   *** Please sign C-Chain ImportTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.C().Signer().SignAtomic(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}

func buildAndSignCChainExportTx(wallet *primary.Wallet, to *secp256k1fx.OutputOwners, destinationChainID ids.ID) (*atomic.Tx, error) {
	baseFee := big.NewInt(25000000000)
	unsignedTx, err := wallet.C().Builder().NewExportTx(
		destinationChainID,
		[]*secp256k1fx.TransferOutput{
			{
				Amt:          500000000,
				OutputOwners: *to,
			},
		},
		baseFee,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	tx := &atomic.Tx{UnsignedAtomicTx: unsignedTx}

	fmt.Println("   *** Please sign C-Chain ExportTx on your Ledger device ***")
	signCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := wallet.C().Signer().SignAtomic(signCtx, tx); err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return tx, nil
}
