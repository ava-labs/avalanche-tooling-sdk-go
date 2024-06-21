// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/host"
)

func main() {
	ctx := context.Background()
	// Get the default cloud parameters for AWS
	cp, err := host.GetDefaultCloudParams(ctx, host.AWSCloud)
	// Set the cloud parameters for AWS non provided by the default
	// Please set your own values for the following fields
	cp.AWSProfile = "default"
	cp.AWSSecurityGroupID = "sg-0e198c427f8f0616b"
	cp.AWSKeyPair = "default"
	if err != nil {
		panic(err)
	}
	// Create a new host instance. Count is 1 so only one host will be created
	hosts, err := host.CreateInstanceList(ctx, *cp, 1)
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
	}
}
