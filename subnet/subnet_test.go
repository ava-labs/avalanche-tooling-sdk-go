// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
)

func getDefaultSubnetEVMGenesis() SubnetParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("INITIAL_ALLOCATION_ADDRESS")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	return SubnetParams{
		SubnetEVM: &SubnetEVMParams{
			ChainID:     big.NewInt(123456),
			FeeConfig:   vm.StarterFeeConfig,
			Allocation:  allocation,
			Precompiles: params.Precompiles{},
		},
		Name: "TestSubnet",
	}
}

func TestSubnetDeploy(t *testing.T) {
	require := require.New(t)
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, err := New(&subnetParams)
	require.NoError(err)
	network := avalanche.FujiNetwork()

	keychain, err := keychain.NewKeychain(network, "KEY_PATH", nil)
	require.NoError(err)

	controlKeys := keychain.Addresses().List()
	subnetAuthKeys := keychain.Addresses().List()
	threshold := 1
	newSubnet.SetSubnetControlParams(controlKeys, uint32(threshold))
	wallet, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	require.NoError(err)
	deploySubnetTx, err := newSubnet.CreateSubnetTx(wallet)
	require.NoError(err)
	subnetID, err := newSubnet.Commit(*deploySubnetTx, wallet, true)
	require.NoError(err)
	fmt.Printf("subnetID %s \n", subnetID.String())
	time.Sleep(2 * time.Second)
	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(wallet)
	require.NoError(err)
	blockchainID, err := newSubnet.Commit(*deployChainTx, wallet, true)
	require.NoError(err)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func TestSubnetDeployMultiSig(t *testing.T) {
	require := require.New(t)
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, _ := New(&subnetParams)
	network := avalanche.FujiNetwork()

	keychainA, err := keychain.NewKeychain(network, "KEY_PATH_A", nil)
	require.NoError(err)
	keychainB, err := keychain.NewKeychain(network, "KEY_PATH_B", nil)
	require.NoError(err)
	keychainC, err := keychain.NewKeychain(network, "KEY_PATH_C", nil)
	require.NoError(err)

	controlKeys := []ids.ShortID{}
	controlKeys = append(controlKeys, keychainA.Addresses().List()[0])
	controlKeys = append(controlKeys, keychainB.Addresses().List()[0])
	controlKeys = append(controlKeys, keychainC.Addresses().List()[0])

	subnetAuthKeys := []ids.ShortID{}
	subnetAuthKeys = append(subnetAuthKeys, keychainA.Addresses().List()[0])
	subnetAuthKeys = append(subnetAuthKeys, keychainB.Addresses().List()[0])
	threshold := 2
	newSubnet.SetSubnetControlParams(controlKeys, uint32(threshold))

	walletA, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainA.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	require.NoError(err)

	deploySubnetTx, err := newSubnet.CreateSubnetTx(walletA)
	require.NoError(err)
	subnetID, err := newSubnet.Commit(*deploySubnetTx, walletA, true)
	require.NoError(err)
	fmt.Printf("subnetID %s \n", subnetID.String())

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)
	// first signature of CreateChainTx using keychain A
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	require.NoError(err)

	// include subnetID in PChainTxsToFetch when creating second wallet
	walletB, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainB.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	require.NoError(err)

	// second signature using keychain B
	err = walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx)
	require.NoError(err)

	// since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, err := newSubnet.Commit(*deployChainTx, walletA, true)
	require.NoError(err)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func TestSubnetDeployLedger(t *testing.T) {
	require := require.New(t)
	subnetParams := getDefaultSubnetEVMGenesis()
	newSubnet, err := New(&subnetParams)
	require.NoError(err)
	network := avalanche.FujiNetwork()

	ledgerInfo := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fujixxxxxxxxx"},
	}
	keychainA, err := keychain.NewKeychain(network, "", &ledgerInfo)
	require.NoError(err)

	addressesIDs, err := address.ParseToIDs([]string{"P-fujiyyyyyyyy"})
	require.NoError(err)
	controlKeys := addressesIDs
	subnetAuthKeys := addressesIDs
	threshold := 1

	newSubnet.SetSubnetControlParams(controlKeys, uint32(threshold))

	walletA, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainA.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)

	require.NoError(err)
	deploySubnetTx, err := newSubnet.CreateSubnetTx(walletA)
	require.NoError(err)
	subnetID, err := newSubnet.Commit(*deploySubnetTx, walletA, true)
	require.NoError(err)
	fmt.Printf("subnetID %s \n", subnetID.String())

	time.Sleep(2 * time.Second)

	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)
	deployChainTx, err := newSubnet.CreateBlockchainTx(walletA)
	require.NoError(err)

	ledgerInfoB := keychain.LedgerParams{
		LedgerAddresses: []string{"P-fujiyyyyyyyy"},
	}
	err = keychainA.Ledger.LedgerDevice.Disconnect()
	require.NoError(err)

	keychainB, err := keychain.NewKeychain(network, "", &ledgerInfoB)
	require.NoError(err)

	walletB, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychainB.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	require.NoError(err)

	// second signature
	err = walletB.P().Signer().Sign(context.Background(), deployChainTx.PChainTx)
	require.NoError(err)

	blockchainID, err := newSubnet.Commit(*deployChainTx, walletB, true)
	require.NoError(err)

	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func TestValidateSubnet(t *testing.T) {
	ctx := context.Background()
	cp, err := node.GetDefaultCloudParams(ctx, node.AWSCloud)
	if err != nil {
		panic(err)
	}

	subnetParams := SubnetParams{
		GenesisFilePath: "/Users/raymondsukanto/.avalanche-cli/subnets/sdkSubnetNew/genesis.json",
		Name:            "sdkSubnetNew",
	}

	newSubnet, err := New(&subnetParams)
	if err != nil {
		panic(err)
	}

	node := node.Node{
		// NodeID is Avalanche Node ID of the node
		NodeID: "NodeID-Mb3AwcUpWysCWLP6mSpzzJVgYawJWzPHu",
		// IP address of the node
		IP: "18.144.79.215",
		// SSH configuration for the node
		SSHConfig: node.SSHConfig{
			User:           constants.RemoteHostUser,
			PrivateKeyPath: "/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem",
		},
		// Cloud is the cloud service that the node is on
		Cloud: node.AWSCloud,
		// CloudConfig is the cloud specific configuration for the node
		CloudConfig: *cp,
		// Role of the node can be 	Validator, API, AWMRelayer, Loadtest, or Monitor
		Roles: []node.SupportedRole{node.Validator},
	}

	subnetIDsToValidate := []string{newSubnet.SubnetID.String()}
	fmt.Printf("Reconfiguring node %s to track subnet %s\n", node.NodeID, subnetIDsToValidate)
	if err := node.SyncSubnets(subnetIDsToValidate); err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)

	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk", nil)
	if err != nil {
		panic(err)
	}

	subnetID, err := ids.FromString("2VsqBt64W9qayKttmGTiAmtsQVnp9e9U4gSHF1yuLKHuquck5j")

	_, err = wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: set.Of(subnetID),
		},
	)
	if err != nil {
		panic(err)
	}

	nodeID, err := ids.NodeIDFromString(node.NodeID)
	if err != nil {
		panic(err)
	}

	_ = ValidatorParams{
		NodeID: nodeID,
		// Validate Subnet for 48 hours
		Duration: 48 * time.Hour,
		Weight:   20,
	}
	fmt.Printf("adding subnet validator")

	newSubnet.SetSubnetID(subnetID)
	subnetAuthKeys := keychain.Addresses().List()
	newSubnet.SetSubnetAuthKeys(subnetAuthKeys)

	//addValidatorTx, err := newSubnet.AddValidator(wallet, validator)
	//if err != nil {
	//	panic(err)
	//}
	//txID, err := newSubnet.Commit(*addValidatorTx, wallet, true)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("obtained tx id %s", txID.String())
}
