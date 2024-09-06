// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	"github.com/ava-labs/avalanche-tooling-sdk-go/interchain/relayer"
	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
)

func InterchainAWSExample(
	network avalanche.Network,
	chain1RPC string,
	chain1PK string,
	chain1SubnetID ids.ID,
	chain1BlockchainID ids.ID,
	chain2RPC string,
	chain2PK string,
	chain2SubnetID ids.ID,
	chain2BlockchainID ids.ID,
) error {
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
	cp.AWSConfig.AWSSecurityGroupName = securityGroupName
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

	const (
		awmVersion = "v1.4.0"
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
			CloudParams:       cp,
			Count:             1,
			Roles:             []node.SupportedRole{node.AWMRelayer},
			Network:           avalanche.FujiNetwork(),
			AWMRelayerVersion: awmVersion,
			UseStaticIP:       false,
			SSHPrivateKeyPath: sshPrivateKeyPath,
		})
	if err != nil {
		panic(err)
	}
	if len(hosts) != 1 {
		panic("expected 1 host")
	}
	awmRelayerHost := hosts[0] // single awm-relayer host

	const (
		sshTimeout        = 120 * time.Second
		sshCommandTimeout = 10 * time.Second
	)

	// Wait for the host to be ready (only needs to be done once for newly created nodes)
	fmt.Println("Waiting for SSH shell")
	if err := awmRelayerHost.WaitForSSHShell(sshTimeout); err != nil {
		panic(err)
	}

	fmt.Println("Deploying Interchain Messenger to AWS")
	chain1RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}
	chain2RelayerKey, err := key.NewSoft()
	if err != nil {
		return err
	}
	chain1RegistryAddress, chain1MessengerAddress, chain2RegistryAddress, chain2MessengerAddress, err := SetupICM(
		chain1RPC,
		chain1PK,
		chain2RPC,
		chain2PK,
	)
	if err != nil {
		return err
	}
	// Get default awm-relayer configuration
	relayerConfig, err := awmRelayerHost.GetAMWRelayerConfig()
	if err != nil {
		panic(err)
	}
	// Add blockchain chain1 to the relayer config,
	// setting it both as source and as destination.
	// So the relayer will both listed for new messages in it,
	// and send to it new messages from other blockchains.
	relayer.AddBlockchainToRelayerConfig(
		relayerConfig,
		chain1RPC,
		"",
		chain1SubnetID,
		chain1BlockchainID,
		chain1RegistryAddress,
		chain1MessengerAddress,
		chain1RelayerKey.C(),
		chain1RelayerKey.PrivKeyHex(),
	)
	// Add blockchain chain2 to the relayer config,
	// setting it both as source and as destination.
	// So the relayer will both listed for new messages in it,
	// and send to it new messages from other blockchains.
	relayer.AddBlockchainToRelayerConfig(
		relayerConfig,
		chain2RPC,
		"",
		chain2SubnetID,
		chain2BlockchainID,
		chain2RegistryAddress,
		chain2MessengerAddress,
		chain2RelayerKey.C(),
		chain2RelayerKey.PrivKeyHex(),
	)
	// Set awm-relayer configuration for the host
	if err := awmRelayerHost.SetAMWRelayerConfig(relayerConfig); err != nil {
		panic(err)
	}

	// Fund each relayer key with 10 TOKENs
	// Where TOKEN is the native gas token of each blockchain
	// Assumes that the TOKEN decimals are 18, so, this equals
	// to 1e18 of the smallest gas amount in each chain
	fmt.Printf("Funding relayer keys %s, %s\n", chain1RelayerKey.C(), chain2RelayerKey.C())
	desiredRelayerBalance := big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(10))

	// chain1PK will have a balance 10 native gas tokens on chain.
	if err := relayer.FundRelayer(
		relayerConfig,
		chain1BlockchainID,
		chain1PK,
		nil,
		desiredRelayerBalance,
	); err != nil {
		return err
	}
	// chain2PK will have a balance 10 native gas tokens on chain2
	if err := relayer.FundRelayer(
		relayerConfig,
		chain2BlockchainID,
		chain2PK,
		nil,
		desiredRelayerBalance,
	); err != nil {
		return err
	}

	// send a message from chain1 to chain2
	fmt.Println("Verifying message delivery")
	if err := TestMessageDelivery(
		chain1RPC,
		chain1PK,
		chain1MessengerAddress,
		chain2BlockchainID,
		chain2RPC,
		chain2MessengerAddress,
		[]byte("hello world"),
	); err != nil {
		return err
	}

	fmt.Println("Message successfully delivered")

	return nil
}

func main() {
	chain1RPC := os.Getenv("CHAIN1_RPC")
	chain1PK := os.Getenv("CHAIN1_PK")
	chain1SubnetID, err := ids.FromString(os.Getenv("CHAIN1_SUBNET_ID"))
	if err != nil {
		panic(err)
	}
	chain1BlockchainID, err := ids.FromString(os.Getenv("CHAIN1_BLOCKCHAIN_ID"))
	if err != nil {
		panic(err)
	}
	chain2RPC := os.Getenv("CHAIN2_RPC")
	chain2PK := os.Getenv("CHAIN2_PK")
	chain2SubnetID, err := ids.FromString(os.Getenv("CHAIN2_SUBNET_ID"))
	if err != nil {
		panic(err)
	}
	chain2BlockchainID, err := ids.FromString(os.Getenv("CHAIN2_BLOCKCHAIN_ID"))
	if err != nil {
		panic(err)
	}
	if err := InterchainAWSExample(
		avalanche.FujiNetwork(),
		chain1RPC,
		chain1PK,
		chain1SubnetID,
		chain1BlockchainID,
		chain2RPC,
		chain2PK,
		chain2SubnetID,
		chain2BlockchainID,
	); err != nil {
		panic(err)
	}
}
