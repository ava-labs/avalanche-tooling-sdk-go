// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

func TestCreateNodes(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	// Get the default cloud parameters for AWS
	cp, err := GetDefaultCloudParams(ctx, AWSCloud)
	require.NoError(err)

	securityGroupName := "SECURITY_GROUP_NAME"
	sgID, err := awsAPI.CreateSecurityGroup(ctx, securityGroupName, cp.AWSConfig.AWSProfile, cp.Region)
	require.NoError(err)

	// Set the security group we are using when creating our Avalanche Nodes
	cp.AWSConfig.AWSSecurityGroupID = sgID
	cp.AWSConfig.AWSSecurityGroupName = securityGroupName

	keyPairName := "KEY_PAIR_NAME"
	sshPrivateKeyPath := utils.ExpandHome("PRIVATE_KEY_FILEPATH")
	err = awsAPI.CreateSSHKeyPair(ctx, cp.AWSConfig.AWSProfile, cp.Region, keyPairName, sshPrivateKeyPath)
	require.NoError(err)

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
	hosts, err := CreateNodes(ctx,
		&NodeParams{
			CloudParams:        cp,
			Count:              2,
			Roles:              []SupportedRole{Validator},
			Network:            avalanche.FujiNetwork(),
			AvalancheGoVersion: avalancheGoVersion,
			UseStaticIP:        false,
			SSHPrivateKeyPath:  sshPrivateKeyPath,
		})
	require.NoError(err)

	fmt.Println("Successfully created Avalanche Validators")

	const (
		sshTimeout        = 120 * time.Second
		sshCommandTimeout = 10 * time.Second
	)

	// Examples showing how to run ssh commands on the created nodes
	for _, h := range hosts {
		// Wait for the host to be ready (only needs to be done once for newly created nodes)
		fmt.Println("Waiting for SSH shell")
		if err := h.WaitForSSHShell(sshTimeout); err != nil {
			require.NoError(err)
		}
		fmt.Println("SSH shell ready to execute commands")
		// Run a command on the host
		if output, err := h.Commandf(nil, sshCommandTimeout, "echo 'Hello, %s!'", "World"); err != nil {
			require.NoError(err)
		} else {
			fmt.Println(string(output))
		}
		// sleep for 10 seconds allowing AvalancheGo container to start
		time.Sleep(10 * time.Second)
		// check if avalanchego is running
		if output, err := h.Commandf(nil, sshCommandTimeout, "docker ps"); err != nil {
			require.NoError(err)
		} else {
			fmt.Println(string(output))
		}
	}

	// Create a monitoring node.
	// Monitoring node enables you to have a centralized Grafana Dashboard where you can view
	// metrics relevant to any Validator & API nodes that the monitoring node is linked to as well
	// as a centralized logs for the X/P/C Chain and Subnet logs for the Validator & API nodes.
	// An example on how the dashboard and logs look like can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-aws
	monitoringHosts, err := CreateNodes(ctx,
		&NodeParams{
			CloudParams:       cp,
			Count:             1,
			Roles:             []SupportedRole{Monitor},
			UseStaticIP:       false,
			SSHPrivateKeyPath: sshPrivateKeyPath,
		})
	require.NoError(err)

	fmt.Println("Successfully created monitoring node")
	fmt.Println("Linking monitoring node with Avalanche Validator nodes ...")
	// Link the 2 validator nodes previously created with the monitoring host so that
	// the monitoring host can start tracking the validator nodes metrics and collecting their logs
	err = monitoringHosts[0].MonitorNodes(ctx, hosts, "")
	require.NoError(err)
	fmt.Println("Successfully linked monitoring node with Avalanche Validator nodes")

	fmt.Println("Terminating all created nodes ...")
	// Destroy all created nodes
	for _, h := range hosts {
		err = h.Destroy(ctx)
		require.NoError(err)
	}
	err = monitoringHosts[0].Destroy(ctx)
	require.NoError(err)
	fmt.Println("All nodes terminated")
}
