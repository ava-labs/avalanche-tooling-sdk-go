// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

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

// Create creates a new node.
// If wait is true, this function will block until the node is ready.
func CreateInstanceList(ctx context.Context, cp CloudParams, count int) ([]Host, error) {
	if err := preCreateCheck(cp, count); err != nil {
		return nil, err
	}
	hosts := make([]Host, 0, count)
	switch cp.Cloud() {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			cp.AWSProfile,
			cp.Region,
		)
		if err != nil {
			return nil, err
		}
		instanceIds, err := ec2Svc.CreateEC2Instances(
			cp.Name,
			count,
			cp.Image,
			cp.InstanceType,
			cp.AWSKeyPair,
			cp.AWSSecurityGroupID,
			cp.AWSVolumeIOPS,
			cp.AWSVolumeThroughput,
			cp.AWSVolumeType,
			cp.AWSVolumeSize,
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
			hosts = append(hosts, Host{
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
		return hosts, nil
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			cp.GCPProject,
			cp.GCPCredentials,
		)
		if err != nil {
			return nil, err
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
		if err != nil {
			return nil, err
		}
		if len(computeInstances) != count {
			return nil, fmt.Errorf("failed to create all instances. Expected %d, got %d", count, len(computeInstances))
		}
		for _, computeInstance := range computeInstances {
			hosts = append(hosts, Host{
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
	return hosts, nil
}
