// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/ava-labs/avalanche-cli/pkg/constants"
	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
)

type CloudParams struct {
	// Name of the node
	Name string

	// Region to use for the node
	Region string

	// Image to use for the node
	Image string

	// Instance type to use for the node
	InstanceType string

	// Static IP to use for the node
	StaticIP string

	// AWS specific configuration

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

	//GCP zone to use for the node
	GCPZone string
}

// New returns a new CloudParams with
func GetDefaultCloudParams(ctx context.Context, cloud SupportedCloud) (*CloudParams, error) {
	// make sure that CloudParams is initialized with default values
	switch cloud {
	case AWSCloud:
		cp := &CloudParams{
			Name:             "avalanche-tooling-sdk-go",
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
		imageID, err := awsSvc.GetUbuntuAMIID(arch, constants.UbuntuVersionLTS)
		if err != nil {
			return nil, err
		}
		cp.Image = imageID
		return cp, nil
	case GCPCloud:
		projectName, err := getDefaultProjectNameFromGCPCredentials(constants.GCPDefaultAuthKeyPath)
		if err != nil {
			return nil, err
		}
		cp := &CloudParams{
			Name:           "avalanche-tooling-sdk-go",
			GCPProject:     projectName,
			GCPCredentials: constants.GCPDefaultAuthKeyPath,
			GCPNetwork:     "avalanche-tooling-sdk-go-us-east1",
			GCPZone:        "us-east1-b",
			Region:         "us-east1",
			InstanceType:   constants.GCPDefaultInstanceType,
			StaticIP:       "",
		}
		gcpSvc, err := gcpAPI.NewGcpCloud(ctx, cp.GCPProject, cp.GCPCredentials)
		if err != nil {
			return nil, err
		}
		imageID, err := gcpSvc.GetUbuntuimageID()
		if err != nil {
			return nil, err
		}
		cp.Image = imageID
		return cp, nil
	default:
		return nil, fmt.Errorf("unsupported cloud")
	}
}

// Validate checks that the CloudParams are valid for deployment
func (cp *CloudParams) Validate() error {
	// common checks
	if cp.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cp.Region == "" {
		return fmt.Errorf("region is required")
	}
	if cp.Image == "" {
		return fmt.Errorf("image is required")
	}
	if cp.InstanceType == "" {
		return fmt.Errorf("instance type is required")
	}
	cloud, err := cp.Cloud()
	if err != nil {
		return err
	}
	switch cloud {
	case AWSCloud:
		if cp.AWSProfile == "" {
			return fmt.Errorf("AWS profile is required")
		}
		if cp.AWSSecurityGroup == "" {
			return fmt.Errorf("AWS security group is required")
		}
		if cp.AWSVolumeSize < 0 {
			return fmt.Errorf("AWS volume size must be positive")
		}
		if cp.AWSVolumeType == "" {
			return fmt.Errorf("AWS volume type is required")
		}
		if cp.AWSVolumeIOPS < 0 {
			return fmt.Errorf("AWS volume IOPS must be positive")
		}
		if cp.AWSVolumeThroughput < 0 {
			return fmt.Errorf("AWS volume throughput must be positive")
		}
	case GCPCloud:
		if cp.GCPNetwork == "" {
			return fmt.Errorf("GCP network is required")
		}
		if cp.GCPProject == "" {
			return fmt.Errorf("GCP project is required")
		}
		if cp.GCPCredentials == "" {
			return fmt.Errorf("GCP credentials is required")
		}
		if cp.GCPZone == "" {
			return fmt.Errorf("GCP zone is required")
		}
		if !strings.HasPrefix(cp.GCPZone, cp.Region) {
			return fmt.Errorf("GCP zone must be in the region %s", cp.Region)
		}
	default:
		return fmt.Errorf("unsupported cloud")
	}
	return nil
}

// Cloud returns the SupportedCloud for the CloudParams
func (cp *CloudParams) Cloud() (SupportedCloud, error) {
	if cp.AWSProfile != "" {
		return AWSCloud, nil
	} else if cp.GCPProject != "" || cp.GCPCredentials != "" {
		return GCPCloud, nil
	} else {
		return Unknown, fmt.Errorf("cloud not specified")
	}
}
