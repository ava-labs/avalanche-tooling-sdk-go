// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
)

// Destroy destroys a node.
func (node *Node) Destroy(ctx context.Context) error {
	switch node.Cloud {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			node.CloudConfig.(CloudParams).AWSProfile,
			node.CloudConfig.(CloudParams).Region,
		)
		if err != nil {
			return err
		}
		return ec2Svc.DestroyAWSNode(node.ID)
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			node.CloudConfig.(CloudParams).GCPProject,
			node.CloudConfig.(CloudParams).GCPCredentials,
		)
		if err != nil {
			return err
		}
		return gcpSvc.DestroyGCPNode(node.CloudConfig.(CloudParams).Region, node.ID)
	default:
		return fmt.Errorf("unsupported cloud type: %s", node.Cloud.String())
	}
}
