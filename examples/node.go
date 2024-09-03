// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

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
		avalancheGoVersion = "v1.11.8"
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
			CloudParams:        cp,
			Count:              2,
			Roles:              []node.SupportedRole{node.Validator},
			Network:            avalanche.FujiNetwork(),
			AvalancheGoVersion: avalancheGoVersion,
			UseStaticIP:        false,
			SSHPrivateKeyPath:  sshPrivateKeyPath,
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

	// Destroy all created nodes
	for _, h := range hosts {
		err = h.Destroy(ctx)
		if err != nil {
			panic(err)
		}
	}
	err = monitoringHosts[0].Destroy(ctx)
	if err != nil {
		panic(err)
	}
}
