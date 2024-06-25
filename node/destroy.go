// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

import (
	"context"
	"fmt"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
)

// Destroy destroys a node.
func (h *Host) Destroy(ctx context.Context) error {
	switch h.Cloud {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			h.CloudConfig.AWSProfile,
			h.CloudConfig.Region,
		)
		if err != nil {
			return err
		}
		return ec2Svc.DestroyAWSNode(h.NodeID)
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			h.CloudConfig.GCPProject,
			h.CloudConfig.GCPCredentials,
		)
		if err != nil {
			return err
		}
		return gcpSvc.DestroyGCPNode(h.CloudConfig.Region, h.NodeID)
	default:
		return fmt.Errorf("unsupported cloud type: %s", h.Cloud.String())
	}
}
