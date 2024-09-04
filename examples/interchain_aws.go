// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

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

func CreateAWMRelayerNode() {
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
		// check if awm-relayer is running
		if output, err := h.Commandf(nil, sshCommandTimeout, "docker ps"); err != nil {
			panic(err)
		} else {
			fmt.Println(string(output))
		}
	}

	for _, h := range hosts {
		fmt.Println("Get awm-relayer configuratuin")
		awmConfig, err := h.GetAMWRelayerConfig()
		if err != nil {
			panic(err)
		}
		fmt.Println(awmConfig)
	}
}
