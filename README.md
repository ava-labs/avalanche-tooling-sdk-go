# avalanche-tooling-sdk-go

The official Avalanche Tooling Go SDK library.

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
object that will be used to create CreateSubnetTx and CreateChainTx and how to commit these 
transactions on chain.

```go
package examples

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
	newSubnet.SetDeployParams(controlKeys, subnetAuthKeys, uint32(threshold))

	wallet, _ := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)

	deploySubnetTx, _ := newSubnet.CreateSubnetTx(wallet)
	subnetID, _ := wallet.Commit(*deploySubnetTx, true)
	fmt.Printf("subnetID %s \n", subnetID.String())
	newSubnet.SubnetID = subnetID

	// we need to wait to allow the transaction to reach other nodes in Fuji
	time.Sleep(2 * time.Second)

	deployChainTx, _ := newSubnet.CreateBlockchainTx(wallet)
	// since we are using the fee paying key as control key too, we can commit the transaction
	// on chain immediately since the number of signatures has been reached
	blockchainID, _ := wallet.Commit(*deployChainTx, true)
	fmt.Printf("blockchainID %s \n", blockchainID.String())
}

func getDefaultSubnetEVMGenesis() subnet.SubnetParams {
	allocation := core.GenesisAlloc{}
	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
	allocation[common.HexToAddress("INITIAL_ALLOCATION_ADDRESS")] = core.GenesisAccount{
		Balance: defaultAmount,
	}
	teleporterInfo := &teleporter.Info{
		Version:                  "v1.0.0",
		FundedAddress:            "0x6e76EEf73Bcb65BCCd16d628Eb0B696552c53E4e",
		FundedBalance:            big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(600)),
		MessengerDeployerAddress: "0x618FEdD9A45a8C456812ecAAE70C671c6249DfaC",
		RelayerAddress:           "0x2A20d1623ce3e90Ec5854c84E508B8af065C059d",
	}
	return subnet.SubnetParams{
		SubnetEVM: &subnet.SubnetEVMParams{
			EnableWarp:     true,
			ChainID:        big.NewInt(123456),
			FeeConfig:      vm.StarterFeeConfig,
			Allocation:     allocation,
			Precompiles:    params.Precompiles{},
			TeleporterInfo: teleporterInfo,
		},
		Name: "TestSubnet",
	}
}
```