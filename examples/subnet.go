// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	"math/big"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
)

func getDefaultSubnetEVMGenesis() subnet.SubnetParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("INITIAL_ALLOCATION_ADDRESS")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	return subnet.SubnetParams{
		SubnetEVM: &subnet.SubnetEVMParams{
			ChainID:     big.NewInt(123456),
			FeeConfig:   vm.StarterFeeConfig,
			Allocation:  allocation,
			Precompiles: params.Precompiles{},
		},
		Name: "TestSubnet",
	}
}

func DeploySubnet() {
	subnetParams := getDefaultSubnetEVMGenesis()
	// Create new Subnet EVM genesis
	newSubnet, _ := subnet.New(&subnetParams)

	network := avalanche.FujiNetwork()

	// Key that will be used for paying the transaction fees of CreateSubnetTx and CreateChainTx
	// NewKeychain will generate a new key pair in the provided path if no .pk file currently
	// exists in the provided path
	keychain, _ := keychain.NewKeychain(network, "KEY_PATH", nil)

	// In this example, we are using the fee-paying key generated above also as control key
	// and subnet auth key

	// Control keys are a list of keys that are permitted to make changes to a Subnet
	// such as creating a blockchain in the Subnet and adding validators to the Subnet
	controlKeys := keychain.Addresses().List()

	// Subnet auth keys are a subset of control keys that will be used to sign transactions that
	// modify a Subnet (such as creating a blockchain in the Subnet and adding validators to the
	// Subnet)
	//
	// Number of keys in subnetAuthKeys has to be equal to the threshold value provided during
	// CreateSubnetTx.
	//
	// All keys in subnetAuthKeys have to sign the transaction before the transaction
	subnetAuthKeys := keychain.Addresses().List()
	threshold := 1
	newSubnet.SetSubnetControlParams(controlKeys, uint32(threshold))

	wallet, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)

	// Build and Sign CreateSubnetTx with our fee paying key
	deploySubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	// Commit our CreateSubnetTx on chain
	subnetID, _ := newSubnet.Commit(*deploySubnetTx, wallet, true)
	fmt.Printf("subnetID %s \n", subnetID.String())

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)
	// Build and Sign CreateChainTx with our fee paying key (which is also our subnet auth key)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	// Commit our CreateChainTx on chain
	// Since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, _ := newSubnet.Commit(*deployChainTx, wallet, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func DeploySubnetWithLedger() {
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := subnet.New(&subnetParams)
	network := avalanche.FujiNetwork()

	// Create keychain with a specific Ledger address. More than 1 address can be used.
	//
	// Alternatively, keychain can also be created from Ledger without specifying any Ledger address
	// by stating the amount of AVAX required to pay for transaction fees. Keychain SDK will
	// then look through all indexes of all addresses in the Ledger until sufficient AVAX balance
	// is reached. For example:
	//
	// fee := network.GenesisParams().CreateBlockchainTxFee + network.GenesisParams().CreateSubnetTxFee
	// ledgerInfo := keychain.LedgerParams{
	//	 RequiredFunds: fee,
	// }
	//
	// To view Ledger addresses and their balances, you can use Avalanche CLI and use the command
	// avalanche key list --ledger [0,1,2,3,4]
	// The example command above will list the first five addresses in your Ledger
	//
	// To transfer funds between addresses in Ledger, refer to https://docs.avax.network/tooling/cli-transfer-funds/how-to-transfer-funds
	ledgerInfo := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fujixxxxxxxxx"},
	}

	// Here we are creating keychain A which will be used as fee-paying key for CreateSubnetTx
	// and CreateChainTx
	keychainA, _ := keychain.NewKeychain(network, "", &ledgerInfo)
	walletA, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainA.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)

	// In this example, we are using a key different from fee-paying key generated above
	// as control key and subnet auth key
	addressesIDs, _ := address.ParseToIDs([]string{"P-fujiyyyyyyyy"})
	controlKeys := addressesIDs
	subnetAuthKeys := addressesIDs
	threshold := 1
	newSubnet.SetSubnetControlParams(controlKeys, uint32(threshold))

	// Pay and Sign CreateSubnet Tx with fee paying key A using Ledger
	deploySubnetTx, _ := newSubnet.CreateSubnetTx(walletA)
	subnetID, _ := newSubnet.Commit(*deploySubnetTx, walletA, true)
	fmt.Printf("subnetID %s \n", subnetID.String())

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)

	// Pay and sign CreateChain Tx with fee paying key A using Ledger
	deployChainTx, _ := newSubnet.CreateBlockchainTx(walletA)

	// We have to first disconnect Ledger to avoid errors when signing with our subnet auth
	// keys later
	_ = keychainA.Ledger.LedgerDevice.Disconnect()

	// Here we are creating keychain B using the Ledger address that was used as the control key and
	// subnet auth key for our subnet.
	ledgerInfoB := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fujiyyyyyyyy"},
	}
	keychainB, _ := keychain.NewKeychain(network, "", &ledgerInfoB)

	// include subnetID in PChainTxsToFetch when creating second wallet
	walletB, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainB.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)

	// Sign with our Subnet auth key
	_ = walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx)

	// Now that the number of signatures have been met, we can commit our transaction
	// on chain
	blockchainID, _ := newSubnet.Commit(*deployChainTx, walletB, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}
