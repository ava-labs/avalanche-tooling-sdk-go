// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

import (
	"context"
	"fmt"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
)

// Create creates a new node.
// If wait is true, this function will block until the node is ready.
func CreateInstance(ctx context.Context, cp CloudParams) (Host, error) {
	if err := cp.Validate(); err != nil {
		return Host{}, err
	}
	switch cp.Cloud() {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			cp.AWSProfile,
			cp.Region,
		)
		if err != nil {
			return Host{}, err
		}
		instanceIds, err := ec2Svc.CreateEC2Instances(
			cp.Name,
			1,
			cp.Image,
			cp.InstanceType,
			cp.AWSKeyPair,
			cp.AWSSecurityGroupID,
			cp.AWSVolumeIOPS,
			cp.AWSVolumeThroughput,
			cp.AWSVolumeType,
			cp.AWSVolumeSize,
		)
		if err != nil || len(instanceIds) == 0 {
			return Host{}, err
		}
		return Host{
			NodeID:      instanceIds[0],
			IP:          "",
			Cloud:       cp.Cloud(),
			CloudConfig: cp,
			Roles:       nil,
		}, nil
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			cp.GCPProject,
			cp.GCPCredentials,
		)
		if err != nil {
			return Host{}, err
		}
		computeInstances, err := gcpSvc.SetupInstances(
			cp.Name,
			cp.GCPZone,
			cp.GCPNetwork,
			cp.GCPSSHKey,
			cp.Image,
			cp.Name,
			cp.InstanceType,
			[]string{cp.StaticIP},
			1,
			cp.GCPVolumeSize,
		)
		if err != nil || len(computeInstances) == 0 {
			return Host{}, err
		}
		return Host{
			NodeID:      computeInstances[0].Name,
			IP:          computeInstances[0].NetworkInterfaces[0].NetworkIP,
			Cloud:       cp.Cloud(),
			CloudConfig: cp,
			Roles:       nil,
		}, nil
	default:
		return Host{}, fmt.Errorf("unsupported cloud")
	}
}
