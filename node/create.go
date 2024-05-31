// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
)

// Create creates a new node.
// If wait is true, this function will block until the node is ready.
func CreateInstance(ctx context.Context, cp CloudParams) (Node, error) {
	if err := cp.Validate(); err != nil {
		return Node{}, err
	}
	switch cp.Cloud {
	case AWSCloud:
		ec2Svc, err:= awsAPI.NewAwsCloud(
			ctx,
			cp.AWSProfile,
			cp.Region,
		)
		if err != nil {
			return Node{}, err
		}
		instanceIds, err := ec2Svc.CreateEC2Instances(
			cp.Name,
			1,
			cp.Image,
			cp.InstanceType,
			cp.InstanceType,
			cp.KeyName,
			cp.SecurityGroup,
			cp.AWSVolumeIOPS,
			cp.AWSVolumeThroughput,
			cp/cp.AWSVolumeType,
			cp.AWSVolumeSize,
		)
		if err != nil || len(instanceIds) == 0{
			return Node{}, err
		}
		return Node{
			ID: instanceID[0],
			IP: "",
			Cloud: cp.Cloud,
			CloudConfig: cp,
			Roles: nil,
		}, nil
		case GCPCloud:
			gcpSvc,err := gcpAPI.NewGCPCloud(
				ctx,
				cp.GCPProject,
				cp.GCPCredentials,
			)
			if err != nil {
				return Node{}, err
			}
			instanceIds, err := gcpSvc.SetupInstances(
				cp.Name,



}
