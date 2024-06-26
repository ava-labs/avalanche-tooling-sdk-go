// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
)

func main() {
	ctx := context.Background()
	// Get the default cloud parameters for AWS
	cp, err := node.GetDefaultCloudParams(ctx, node.AWSCloud)
	// Set the cloud parameters for AWS non provided by the default
	// Please set your own values for the following fields
	cp.AWSConfig.AWSProfile = "default"
	cp.AWSConfig.AWSSecurityGroupID = "sg-0e198c427f8f0616b"
	cp.AWSConfig.AWSKeyPair = "artur-us-east-1-avalanche-cli"
	if err != nil {
		panic(err)
	}
	// Create a new host instance. Count is 1 so only one host will be created
	const (
		avalanchegoVersion  = "v1.11.8"
		avalancheCliVersion = "v1.6.2"
	)
	hosts, err := node.CreateNodes(ctx,
		&node.NodeParams{
			CloudParams:         cp,
			Count:               2,
			Roles:               []node.SupportedRole{node.Validator},
			NetworkID:           "fuji",
			AvalancheGoVersion:  avalanchegoVersion,
			AvalancheCliVersion: avalancheCliVersion,
			UseStaticIP:         false,
		})
	if err != nil {
		panic(err)
	}

	const (
		sshTimeout        = 120 * time.Second
		sshCommandTimeout = 10 * time.Second
	)
	for _, h := range hosts {
		// Wait for the host to be ready
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
		// sleep for 10 seconds allowing avalancghego container to start
		time.Sleep(10 * time.Second)
		// check if avalanchego is running
		if output, err := h.Commandf(nil, sshCommandTimeout, "docker ps"); err != nil {
			panic(err)
		} else {
			fmt.Println(string(output))
		}
	}
	fmt.Println("About to create a monitoring node")
	// Create a monitoring node
	monitoringHosts, err := node.CreateNodes(ctx,
		&node.NodeParams{
			CloudParams:         cp,
			Count:               1,
			Roles:               []node.SupportedRole{node.Monitor},
			NetworkID:           "",
			AvalancheGoVersion:  "",
			AvalancheCliVersion: "",
			UseStaticIP:         false,
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("Monitoring host SSH shell ready to execute commands")
	// Register nodes with monitoring host
	if err := monitoringHosts[0].RegisterWithMonitoring(hosts, ""); err != nil {
		panic(err)
	}
}
