// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
)

type CloudParams struct {
	CommonParams
	AWSParams
	GCPParams
}

type CommonParams struct {
	// Region to use for the node
	Region string

	// Image to use for the node
	Image string

	// Instance type to use for the node
	InstanceType string

	// Static IP to use for the node
	StaticIP string
}

// AWS Paramsific configuration
type AWSParams struct {
	// AWS profile to use for the node
	Profile string

	// AWS volume size in GB
	VolumeSize int

	// AWS volume type
	VolumeType string

	// AWS volume IOPS
	VolumeIOPS int

	// AWS volume throughput
	VolumeThroughput int

	// AWS security group to use for the node
	SecurityGroup string
}

type GCPParams struct {
	// GCP project to use for the node
	Project string

	// GCP credentials to use for the node
	Credentials string

	// GCP network label to use for the node
	Network string
}

// New returns a new CloudParams with
func GetDefaultValues(ctx context.Context, cloud SupportedCloud) (*CloudParams, error) {
	//make sure that CloudParams is initialized with default values
	switch cloud {
	case AWSCloud:
		cp := &CloudParams{
			Profile:       "default",
			VolumeSize:    1000,
			VolumeType:    "gp3",
			SecurityGroup: "avalanche-tooling-sdk-go-us-east-1",
			Region:        "us-east-1",
			InstanceType:  constants.AWSDefaultInstanceType,
			StaticIP:      "",
		}
		awsSvc, err := awsAPI.NewAwsCloud(ctx, cp.Profile, cp.Region)
		if err != nil {
			return nil, err
		}
		arch, err := awsSvc.GetInstanceTypeArch(cp.InstanceType)
		if err != nil {
			return nil, err
		}
		imageId, error := awsAPI.GetUbuntuAMIID(arch, constants.UbuntuVersionLTS)
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
			Project:      projectName,
			Credentials:  constants.GCPDefaultAuthKeyPath,
			Network:      "avalanche-tooling-sdk-go-us-east1",
			Region:       "us-east1",
			InstanceType: constants.GCPDefaultInstanceType,
			StaticIP:     "",
		}
		gcpSvc, err := gcpAPI.NewGcpCloud(ctx, cp.Project, cp.Credentials)
		if err != nil {
			return nil, err
		}
		imageID, err := gcpCloud.GetUbuntuImageID()
		if err != nil {
			return nil, err
		}
		cp.Image = imageID
		return cp, nil
	default:
		return nil, fmt.Errorf("unsupported cloud: %s", cloud)
	}
}
