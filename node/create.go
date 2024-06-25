// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// preCreateCheck checks if the cloud parameters are valid.
func preCreateCheck(cp CloudParams, count int) error {
	if count < 1 {
		return fmt.Errorf("count must be at least 1")
	}
	if err := cp.Validate(); err != nil {
		return err
	}
	return nil
}

// CreateNodes launches the specified number of instances on the selected cloud platform.

func CreateNodes(ctx context.Context, cp CloudParams, count int) ([]Node, error) {
	if err := preCreateCheck(cp, count); err != nil {
		return nil, err
	}
	nodes := make([]Node, 0, count)
	switch cp.Cloud() {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			cp.AWSConfig.AWSProfile,
			cp.Region,
		)
		if err != nil {
			return nil, err
		}
		instanceIds, err := ec2Svc.CreateEC2Instances(
			count,
			cp.ImageID,
			cp.InstanceType,
			cp.AWSConfig.AWSKeyPair,
			cp.AWSConfig.AWSSecurityGroupID,
			cp.AWSConfig.AWSVolumeIOPS,
			cp.AWSConfig.AWSVolumeThroughput,
			cp.AWSConfig.AWSVolumeType,
			cp.AWSConfig.AWSVolumeSize,
		)
		if err != nil {
			return nil, err
		}
		if len(instanceIds) != count {
			return nil, fmt.Errorf("failed to create all instances. Expected %d, got %d", count, len(instanceIds))
		}
		if err := ec2Svc.WaitForEC2Instances(instanceIds, types.InstanceStateNameRunning); err != nil {
			return nil, err
		}
		instanceEIPMap, err := ec2Svc.GetInstancePublicIPs(instanceIds)
		if err != nil {
			return nil, err
		}
		for _, instanceID := range instanceIds {
			nodes = append(nodes, Node{
				NodeID:      instanceID,
				IP:          instanceEIPMap[instanceID],
				Cloud:       cp.Cloud(),
				CloudConfig: cp,
				SSHConfig: SSHConfig{
					User: constants.AnsibleSSHUser,
				},
				Roles: nil,
			})
		}
		return nodes, nil
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			cp.GCPConfig.GCPProject,
			cp.GCPConfig.GCPCredentials,
		)
		if err != nil {
			return nil, err
		}
		computeInstances, err := gcpSvc.SetupInstances(
			cp.GCPConfig.GCPZone,
			cp.GCPConfig.GCPNetwork,
			cp.GCPConfig.GCPSSHKey,
			cp.ImageID,
			cp.InstanceType,
			[]string{},
			1,
			cp.GCPConfig.GCPVolumeSize,
		)
		if err != nil {
			return nil, err
		}
		if len(computeInstances) != count {
			return nil, fmt.Errorf("failed to create all instances. Expected %d, got %d", count, len(computeInstances))
		}
		for _, computeInstance := range computeInstances {
			nodes = append(nodes, Node{
				NodeID:      computeInstance.Name,
				IP:          computeInstance.NetworkInterfaces[0].NetworkIP,
				Cloud:       cp.Cloud(),
				CloudConfig: cp,
				SSHConfig: SSHConfig{
					User: constants.AnsibleSSHUser,
				},
				Roles: nil,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported cloud")
	}
	return nodes, nil
}
