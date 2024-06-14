# Avalanche Tooling Go SDK

The official Avalanche Tooling Go SDK library.

*** Please note that this SDK is in experimental mode, major changes to the SDK are to be expected
in between releases ***

Current version (v0.1.0) currently only supports Create Subnet and Create Blockchain in a 
Subnet in Fuji / Mainnet. 

Currently, only stored keys are supported for transaction building and signing, ledger support is 
coming soon.

Future SDK releases will contain the following features: 
- Additional Subnet SDK features (i.e. Adding Validators to a Subnet, Custom Subnets, Exporting Subnets)
- Ledger Support 
- Teleporter Support
- Nodes SDK (Creating and setting up Avalanche Validator Nodes and API Nodes )

## Getting Started

### Installing
Use `go get` to retrieve the SDK to add it to your project's Go module dependencies.

	go get github.com/ava-labs/avalanche-tooling-sdk-go

To update the SDK use `go get -u` to retrieve the latest version of the SDK.

	go get -u github.com/ava-labs/avalanche-tooling-sdk-go

## Quick Examples

### Subnet SDK Example

This example shows how to create a Subnet Genesis, deploy the Subnet into Fuji Network and create
a blockchain in the Subnet. 

This examples also shows how to create a key pair to pay for transactions, how to create a Wallet
object that will be used to build and sign CreateSubnetTx and CreateChainTx and how to commit these 
transactions on chain.

```go
package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"
	"github.com/ava-labs/avalanche-tooling-sdk-go/teleporter"
	"github.com/ava-labs/avalanche-tooling-sdk-go/vm"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

// Creates a Subnet Genesis, deploys the Subnet into Fuji Network using CreateSubnetTx
// and creates a blockchain in the Subnet using CreateChainTx
func DeploySubnet() {
	subnetParams := getDefaultSubnetEVMGenesis()
	// Create new Subnet EVM genesis
	newSubnet, _ := subnet.New(&subnetParams)

	network := avalanche.FujiNetwork()

	// Key that will be used for paying the transaction fees of CreateSubnetTx and CreateChainTx
	// NewKeychain will generate a new key pair in the provided path if no .pk file currently
	// exists in the provided path
	keychain, _ := keychain.NewKeychain(network, "KEY_PATH")

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
	// can be committed on chain
	subnetAuthKeys := controlKeys
	threshold := 1
	newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))

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

	newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	// Build and Sign CreateChainTx with our fee paying key (which is also our subnet auth key)
	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	// Commit our CreateChainTx on chain
	// Since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, _ := newSubnet.Commit(*deployChainTx, wallet, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

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
```
