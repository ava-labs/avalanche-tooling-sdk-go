# Avalanche Tooling Go SDK

The official Avalanche Tooling Go SDK library.

*** Please note that this SDK is in experimental mode, major changes to the SDK are to be expected
in between releases ***

Current version (v0.2.0) currently supports: 
- Create Subnet and Create Blockchain in a Subnet in Fuji / Mainnet. 
- Create Avalanche Node (Validator / API / Monitoring / Load Test Node) & install all required
dependencies (AvalancheGo, gcc, Promtail, Grafana, etc).

Currently, only stored keys are supported for transaction building and signing, ledger support is 
coming soon.

Future SDK releases will contain the following features (in order of priority): 
- Additional Nodes SDK features (Have Avalanche nodes validate a Subnet & Primary Network)
- Additional Subnet SDK features (i.e. Adding Validators to a Subnet, Custom Subnets)
- Teleporter SDK
- Ledger SDK

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

### Nodes SDK Example

This example shows how to create Avalanche Validator Nodes (SDK function for nodes to start 
validating Primary Network / Subnet will be available in the next Avalanche Tooling SDK release).

This examples also shows how to create an Avalanche Monitoring Node, which enables you to have a 
centralized Grafana Dashboard where you can view metrics relevant to any Validator & API nodes that
the monitoring node is linked to as well as a centralized logs for the X/P/C Chain and Subnet logs 
for the Validator & API nodes. An example on how the dashboard and logs look like can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-aws

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

func CreateNodes() {
	ctx := context.Background()

	// Get the default cloud parameters for AWS
	cp, err := node.GetDefaultCloudParams(ctx, node.AWSCloud)
	if err != nil {
		panic(err)
	}

	securityGroupName := "SECURITY_GROUP_NAME"
	// Create a new security group in AWS if you do not currently have one in the selected
	// AWS region.
	sgID, err := awsAPI.CreateSecurityGroup(ctx, securityGroupName, cp.AWSConfig.AWSProfile, cp.Region)
	if err != nil {
		panic(err)
	}
	// Set the security group we are using when creating our Avalanche Nodes
	cp.AWSConfig.AWSSecurityGroupID = sgID

	keyPairName := "KEY_PAIR_NAME"
	sshPrivateKeyPath := utils.ExpandHome("PRIVATE_KEY_FILEPATH")
	// Create a new AWS SSH key pair if you do not currently have one in your selected AWS region.
	// Note that the created key pair can only be used in the region that it was created in.
	// The private key to the created key pair will be stored in the filepath provided in
	// sshPrivateKeyPath.
	if err := awsAPI.CreateSSHKeyPair(ctx, cp.AWSConfig.AWSProfile, cp.Region, keyPairName, sshPrivateKeyPath); err != nil {
		panic(err)
	}
	// Set the key pair we are using when creating our Avalanche Nodes
	cp.AWSConfig.AWSKeyPair = keyPairName

	// Avalanche-CLI is installed in nodes to enable them to join subnets as validators
	// Avalanche-CLI dependency by Avalanche nodes will be deprecated in the next release
	// of Avalanche Tooling SDK
	const (
		avalancheGoVersion  = "v1.11.8"
	)

	// Create two new Avalanche Validator nodes on Fuji Network on AWS without Elastic IPs
	// attached. Once CreateNodes is completed, the validators will begin bootstrapping process
	// to Primary Network in Fuji Network. Nodes need to finish bootstrapping process
	// before they can validate Avalanche Primary Network / Subnet.
	//
	// SDK function for nodes to start validating Primary Network / Subnet will be available
	// in the next Avalanche Tooling SDK release.
	hosts, err := node.CreateNodes(ctx,
		&node.NodeParams{
			CloudParams:         cp,
			Count:               2,
			Roles:               []node.SupportedRole{node.Validator},
			Network:             avalanche.FujiNetwork(),
			AvalancheGoVersion:  avalancheGoVersion,
			UseStaticIP:         false,
			SSHPrivateKeyPath:   sshPrivateKeyPath,
		})
	if err != nil {
		panic(err)
	}

	const (
		sshTimeout        = 120 * time.Second
		sshCommandTimeout = 10 * time.Second
	)

	// Examples showing how to run ssh commands on the created nodes
	for _, h := range hosts {
		// Wait for the host to be ready (only needs to be done once for newly created nodes)
		fmt.Println("Waiting for SSH shell")
		if err := h.WaitForSSHShell(sshTimeout); err != nil {
			panic(err)
		}
		fmt.Println("SSH shell ready to execute commands")
		// Run a command on the host
		if output, err := h.Commandf(nil, sshCommandTimeout, "echo 'Hello, %s!'", "World"); err != nil {
			panic(err)
		} else {
			fmt.Println(string(output))
		}
		// sleep for 10 seconds allowing AvalancheGo container to start
		time.Sleep(10 * time.Second)
		// check if avalanchego is running
		if output, err := h.Commandf(nil, sshCommandTimeout, "docker ps"); err != nil {
			panic(err)
		} else {
			fmt.Println(string(output))
		}
	}

	// Create a monitoring node.
	// Monitoring node enables you to have a centralized Grafana Dashboard where you can view
	// metrics relevant to any Validator & API nodes that the monitoring node is linked to as well
	// as a centralized logs for the X/P/C Chain and Subnet logs for the Validator & API nodes.
	// An example on how the dashboard and logs look like can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-aws
	monitoringHosts, err := node.CreateNodes(ctx,
		&node.NodeParams{
			CloudParams:       cp,
			Count:             1,
			Roles:             []node.SupportedRole{node.Monitor},
			UseStaticIP:       false,
			SSHPrivateKeyPath: sshPrivateKeyPath,
		})
	if err != nil {
		panic(err)
	}

	// Link the 2 validator nodes previously created with the monitoring host so that
	// the monitoring host can start tracking the validator nodes metrics and collecting their logs
	if err := monitoringHosts[0].MonitorNodes(ctx, hosts, ""); err != nil {
		panic(err)
	}
}
```
