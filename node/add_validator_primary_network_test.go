// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"testing"
	"time"
)

//func TestNodesValidatePrimaryNetwork(_ *testing.T) {
//	ctx := context.Background()
//	// Get the default cloud parameters for AWS
//	cp, err := GetDefaultCloudParams(ctx, AWSCloud)
//	if err != nil {
//		panic(err)
//	}
//
//	//securityGroupName := "SECURITY_GROUP_NAME"
//	securityGroupName := "rs_security_group_sdk_new"
//	//sgID := "sg-024ef9cde0f63fe5a"
//	sgID, err := awsAPI.CreateSecurityGroup(ctx, securityGroupName, cp.AWSConfig.AWSProfile, cp.Region)
//	if err != nil {
//		panic(err)
//	}
//	// Set the security group we are using when creating our Avalanche Nodes
//	cp.AWSConfig.AWSSecurityGroupID = sgID
//	cp.AWSConfig.AWSSecurityGroupName = securityGroupName
//
//	//keyPairName := "KEY_PAIR_NAME"
//	keyPairName := "rs_key_pair_sdk"
//	//sshPrivateKeyPath := utils.ExpandHome("PRIVATE_KEY_FILEPATH")
//	sshPrivateKeyPath := utils.ExpandHome("/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem")
//	//if err := awsAPI.CreateSSHKeyPair(ctx, cp.AWSConfig.AWSProfile, cp.Region, keyPairName, sshPrivateKeyPath); err != nil {
//	//	panic(err)
//	//}
//	// Set the key pair we are using when creating our Avalanche Nodes
//	cp.AWSConfig.AWSKeyPair = keyPairName
//
//	// Avalanche-CLI is installed in nodes to enable them to join subnets as validators
//	// Avalanche-CLI dependency by Avalanche nodes will be deprecated in the next release
//	// of Avalanche Tooling SDK
//	const (
//		avalancheGoVersion = "v1.11.8"
//	)
//
//	// Create two new Avalanche Validator nodes on Fuji Network on AWS without Elastic IPs
//	// attached. Once CreateNodes is completed, the validators will begin bootstrapping process
//	// to Primary Network in Fuji Network. Nodes need to finish bootstrapping process
//	// before they can validate Avalanche Primary Network / Subnet.
//	//
//	// SDK function for nodes to start validating Primary Network / Subnet will be available
//	// in the next Avalanche Tooling SDK release.
//	hosts, err := CreateNodes(ctx,
//		&NodeParams{
//			CloudParams:        cp,
//			Count:              2,
//			Roles:              []SupportedRole{Validator},
//			Network:            avalanche.FujiNetwork(),
//			AvalancheGoVersion: avalancheGoVersion,
//			UseStaticIP:        false,
//			SSHPrivateKeyPath:  sshPrivateKeyPath,
//		})
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println("Successfully created Avalanche Validators")
//
//	const (
//		sshTimeout        = 120 * time.Second
//		sshCommandTimeout = 10 * time.Second
//	)
//
//	// Examples showing how to run ssh commands on the created nodes
//	for _, h := range hosts {
//		// Wait for the host to be ready (only needs to be done once for newly created nodes)
//		fmt.Println("Waiting for SSH shell")
//		if err := h.WaitForSSHShell(sshTimeout); err != nil {
//			panic(err)
//		}
//		fmt.Println("SSH shell ready to execute commands")
//		// Run a command on the host
//		if output, err := h.Commandf(nil, sshCommandTimeout, "echo 'Hello, %s!'", "World"); err != nil {
//			panic(err)
//		} else {
//			fmt.Println(string(output))
//		}
//		// sleep for 10 seconds allowing AvalancheGo container to start
//		time.Sleep(10 * time.Second)
//		// check if avalanchego is running
//		if output, err := h.Commandf(nil, sshCommandTimeout, "docker ps"); err != nil {
//			panic(err)
//		} else {
//			fmt.Println(string(output))
//		}
//	}
//
//	// Create a monitoring node.
//	// Monitoring node enables you to have a centralized Grafana Dashboard where you can view
//	// metrics relevant to any Validator & API nodes that the monitoring node is linked to as well
//	// as a centralized logs for the X/P/C Chain and Subnet logs for the Validator & API nodes.
//	// An example on how the dashboard and logs look like can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-aws
//	monitoringHosts, err := CreateNodes(ctx,
//		&NodeParams{
//			CloudParams:       cp,
//			Count:             1,
//			Roles:             []SupportedRole{Monitor},
//			UseStaticIP:       false,
//			SSHPrivateKeyPath: sshPrivateKeyPath,
//		})
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println("Successfully created monitoring node")
//	fmt.Println("Linking monitoring node with Avalanche Validator nodes ...")
//	// Link the 2 validator nodes previously created with the monitoring host so that
//	// the monitoring host can start tracking the validator nodes metrics and collecting their logs
//	if err := monitoringHosts[0].MonitorNodes(ctx, hosts, ""); err != nil {
//		panic(err)
//	}
//	fmt.Println("Successfully linked monitoring node with Avalanche Validator nodes")
//	//
//	//fmt.Println("Terminating all created nodes ...")
//	//// Destroy all created nodes
//	//for _, h := range hosts {
//	//	err = h.Destroy(ctx)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//	//err = monitoringHosts[0].Destroy(ctx)
//	//if err != nil {
//	//	panic(err)
//	//}
//	//fmt.Println("All nodes terminated")
//}

func TestNodesValidatePrimaryNetwork(_ *testing.T) {
	ctx := context.Background()
	cp, err := GetDefaultCloudParams(ctx, AWSCloud)
	if err != nil {
		panic(err)
	}

	node := Node{
		// NodeID is Avalanche Node ID of the node
		NodeID: "NodeID-Mb3AwcUpWysCWLP6mSpzzJVgYawJWzPHu",
		// IP address of the node
		IP: "18.144.79.215",
		// SSH configuration for the node
		SSHConfig: SSHConfig{
			User:           constants.AnsibleSSHUser,
			PrivateKeyPath: "/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem",
		},
		// Cloud is the cloud service that the node is on
		Cloud: AWSCloud,

		// CloudConfig is the cloud specific configuration for the node
		CloudConfig: *cp,

		Roles: []SupportedRole{Validator},
	}

	//node := Node{
	//	// NodeID is Avalanche Node ID of the node
	//	NodeID: "NodeID-mF1xjcfTNmB3FikUhZg4Lyfpj8htSmZ1",
	//	// IP address of the node
	//	IP: "52.53.210.128",
	//	// SSH configuration for the node
	//	SSHConfig: SSHConfig{
	//		User:           constants.AnsibleSSHUser,
	//		PrivateKeyPath: "/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem",
	//	},
	//	// Cloud is the cloud service that the node is on
	//	Cloud: AWSCloud,
	//
	//	// CloudConfig is the cloud specific configuration for the node
	//	CloudConfig: *cp,
	//
	//	Roles: []SupportedRole{Validator},
	//}

	err = node.ProvideStakingCertAndKey(fmt.Sprintf("/Users/raymondsukanto/.avalanche-cli/nodes/%s", node.NodeID))
	if err != nil {
		panic(err)
	}
	nodeID, err := ids.NodeIDFromString(node.NodeID)
	if err != nil {
		panic(err)
	}
	validator := ValidatorParams{
		NodeID:   nodeID,
		Duration: 48 * time.Hour,
		Weight:   20,
	}
	network := avalanche.FujiNetwork()
	keychain, err := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk", nil)
	if err != nil {
		panic(err)
	}
	wallet, err := wallet.New(
		context.Background(),
		&primary.WalletConfig{
			URI:              network.Endpoint,
			AVAXKeychain:     keychain.Keychain,
			EthKeychain:      secp256k1fx.NewKeychain(),
			PChainTxsToFetch: nil,
		},
	)
	if err != nil {
		panic(err)
	}
	err = node.SetNodeBLSKey(fmt.Sprintf("/Users/raymondsukanto/.avalanche-cli/nodes/%s/signer.key", node.NodeID))
	if err != nil {
		panic(err)
	}
	txID, err := node.AddNodeAsPrimaryNetworkValidator(avalanche.FujiNetwork(), validator, wallet)
	if err != nil {
		panic(err)
	}
	fmt.Printf("obtained tx id %s", txID.String())
}
