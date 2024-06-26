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

// CreateNodes launches the specified number of nodes on the selected cloud platform.
// The role of the node (Avalanche Validator / API / monitoring / load test node) is
// specified through the input CloudParams
//
// Prior to calling CreateNodes, the credentials for AWS / GCP will first need to be set up.
// To set up AWS credentials, more info can be found at https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html#file-format-creds
// To set up GCP credentials, more info can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-gcp#prerequisites
//
// When CreateNodes is used to create Avalanche Validator Nodes, it will:
//   - Launch the specified number of validator nodes on AWS / GCP
//   - Install Docker and have required dependencies (Avalanche Go, gcc, go) installed as Docker
//     images.
//
// Note that currently only Ubuntu Machine Images are supported on CreateNodes
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
