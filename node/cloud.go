// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanche-cli/pkg/constants"
	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
)

type CloudParams struct {
	// Region to use for the node
	Region string

	// Image to use for the node
	Image string

	// Instance type to use for the node
	InstanceType string

	// Static IP to use for the node
	StaticIP string

	// AWS Paramsific configuration

	// AWS profile to use for the node
	AWSProfile string

	// AWS volume size in GB
	AWSVolumeSize int

	// AWS volume type
	AWSVolumeType string

	// AWS volume IOPS
	AWSVolumeIOPS int

	// AWS volume throughput
	AWSVolumeThroughput int

	// AWS security group to use for the node
	AWSSecurityGroup string

	// GCP Specific configuration

	// GCP project to use for the node
	GCPProject string

	// GCP credentials to use for the node
	GCPCredentials string

	// GCP network label to use for the node
	GCPNetwork string
}

// New returns a new CloudParams with
func GetDefaultValues(ctx context.Context, cloud SupportedCloud) (*CloudParams, error) {
	//make sure that CloudParams is initialized with default values
	switch cloud {
	case AWSCloud:
		cp := &CloudParams{
			AWSProfile:       "default",
			AWSVolumeSize:    1000,
			AWSVolumeType:    "gp3",
			AWSSecurityGroup: "avalanche-tooling-sdk-go-us-east-1",
			Region:           "us-east-1",
			InstanceType:     constants.AWSDefaultInstanceType,
			StaticIP:         "",
		}
		awsSvc, err := awsAPI.NewAwsCloud(ctx, cp.AWSProfile, cp.Region)
		if err != nil {
			return nil, err
		}
		arch, err := awsSvc.GetInstanceTypeArch(cp.InstanceType)
		if err != nil {
			return nil, err
		}
		imageId, err := awsSvc.GetUbuntuAMIID(arch, constants.UbuntuVersionLTS)
		if err != nil {
			return nil, err
		}
		cp.Image = imageId
		return cp, nil
	case GCPCloud:
		projectName, err := getDefaultProjectNameFromGCPCredentials(constants.GCPDefaultAuthKeyPath)
		if err != nil {
			return nil, err
		}
		cp := &CloudParams{
			GCPProject:     projectName,
			GCPCredentials: constants.GCPDefaultAuthKeyPath,
			GCPNetwork:     "avalanche-tooling-sdk-go-us-east1",
			Region:         "us-east1",
			InstanceType:   constants.GCPDefaultInstanceType,
			StaticIP:       "",
		}
		gcpClient, err := gcpAPI.NewGCPClient(ctx, cp.GCPProject)
		if err != nil {
			return nil, err
		}
		gcpSvc, err := gcpAPI.NewGcpCloud(ctx, gcpClient, cp.GCPProject)
		if err != nil {
			return nil, err
		}
		imageID, err := gcpSvc.GetUbuntuImageID()
		if err != nil {
			return nil, err
		}
		cp.Image = imageID
		return cp, nil
	default:
		return nil, fmt.Errorf("unsupported cloud: %s", cloud)
	}
}
