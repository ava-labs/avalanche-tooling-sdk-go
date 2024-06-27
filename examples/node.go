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

// createSecurityGroup creates a new security group in AWS using the specified AWS profile and region.
//
// ctx: The context.Context object for the request.
// awsProfile: The AWS profile to use for the request.
// awsRegion: The AWS region to use for the request.
// Returns the ID of the created security group and an error, if any.
func createSecurityGroup(ctx context.Context, awsProfile string, awsRegion string) (string, error) {
	ec2Svc, err := awsAPI.NewAwsCloud(
		ctx,
		awsProfile,
		awsRegion,
	)
	if err != nil {
		return "", err
	}
	// detect user IP address
	userIPAddress, err := utils.GetUserIPAddress()
	if err != nil {
		return "", err
	}
	const securityGroupName = "avalanche-tooling-sdk-go"
	return ec2Svc.SetupSecurityGroup(userIPAddress, securityGroupName)
}

// createSSHKeyPair creates an SSH key pair for AWS.
//
// ctx: The context for the request.
// awsProfile: The AWS profile to use for the request.
// awsRegion: The AWS region to use for the request.
// sshPrivateKeyPath: The path to save the SSH private key.
// Returns an error if unable to create the key pair.
func createSSHKeyPair(ctx context.Context, awsProfile string, awsRegion string, keyPairName string, sshPrivateKeyPath string) error {
	if utils.FileExists(sshPrivateKeyPath) {
		return fmt.Errorf("ssh private key %s already exists", sshPrivateKeyPath)
	}
	ec2Svc, err := awsAPI.NewAwsCloud(
		ctx,
		awsProfile,
		awsRegion,
	)
	if err != nil {
		return err
	}
	return ec2Svc.CreateAndDownloadKeyPair(keyPairName, sshPrivateKeyPath)
}

func CreateNodes() {
	ctx := context.Background()

	// Get the default cloud parameters for AWS
	cp, err := node.GetDefaultCloudParams(ctx, node.AWSCloud)
	if err != nil {
		panic(err)
	}
	// Set the cloud parameters for AWS non provided by the default
	// Please set your own values for the following fields
	// For example cp.AWSConfig.AWSProfile = "default" and so on

	// Create security group or you can provide your own by setting cp.AWSConfig.AWSSecurityGroupID
	sgID, err := createSecurityGroup(ctx, cp.AWSConfig.AWSProfile, cp.Region)
	if err != nil {
		panic(err)
	}
	cp.AWSConfig.AWSSecurityGroupID = sgID

	// Change to your own path
	sshPrivateKey := utils.ExpandHome("~/.ssh/avalanche-tooling-sdk-go.pem")

	// Create aws ssh key pair or you can provide your own by setting cp.AWSConfig.AWSKeyPair.
	// Make sure that ssh private key matches AWSKeyPair
	keyPairName := "avalanche-tooling-sdk-go"
	if err := createSSHKeyPair(ctx, cp.AWSConfig.AWSProfile, cp.Region, keyPairName, sshPrivateKey); err != nil {
		panic(err)
	}
	cp.AWSConfig.AWSKeyPair = keyPairName

	// Create a new host instance. Count is 1 so only one host will be created
	// Set the cloud parameters for AWS non provided by the default
	// Please set your own values for the following fields
	cp.AWSConfig.AWSProfile = "default"
	cp.AWSConfig.AWSSecurityGroupID = "AWS_SECURITY_GROUP_ID"
	cp.AWSConfig.AWSKeyPair = "AWS_KEY_PAIR"
	if err != nil {
		panic(err)
	}
	// Avalanche-CLI is installed in nodes to enable them to join subnets as validators
	// Avalanche-CLI dependency by Avalanche nodes will be deprecated in the next release
	// of Avalanche Tooling SDK

	const (
		avalancheGoVersion  = "v1.11.8"
		avalancheCliVersion = "v1.6.2"
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
			AvalancheCliVersion: avalancheCliVersion,
			UseStaticIP:         false,
			SSHPrivateKey:       sshPrivateKey,
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
			CloudParams: cp,
			Count:       1,
			Roles:       []node.SupportedRole{node.Monitor},
			UseStaticIP: false,
		})
	if err != nil {
		panic(err)
	}

	// Link the 2 validator nodes previously created with the monitoring host so that
	// the monitoring host can start tracking the validator nodes metrics and collecting their logs
	if err := monitoringHosts[0].MonitorNodes(hosts, ""); err != nil {
		panic(err)
	}
}
