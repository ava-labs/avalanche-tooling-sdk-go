// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"strings"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// CloudParams contains the specs of the nodes to be created in AWS / GCP
type CloudParams struct {
	// Region to use for the node
	Region string

	// ImageID is Machine Image ID to use for the node.
	// For example, Machine Image ID for Ubuntu 22.04 LTS (HVM), SSD Volume Type on AWS in
	// us-west-2 region is ami-0cf2b4e024cdb6960 at the time of this writing
	// Note that only Ubuntu Machine Images are supported.
	// To view list of available Machine Images:
	// - AWS: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/finding-an-ami.html
	// - GCP: https://cloud.google.com/compute/docs/images#os-compute-support
	//
	// Avalanche Tooling team publishes our own Ubuntu 20.04 Machine Image called Avalanche-CLI
	// Ubuntu 20.04 Docker on AWS & GCP for both arm64 and amd64 architecture.
	// A benefit to using Avalanche-CLI Ubuntu 20.04 Docker is that it has all the dependencies
	// that an Avalanche Node requires (AvalancheGo, gcc, go, etc), thereby decreasing in massive
	// reduction in time required to provision a node.
	// TODO: add how to get our amiID
	ImageID string

	// Instance type of the node
	// For example c5.2xlarge in AWS
	// For more information about Instance Types:
	// - AWS: https://aws.amazon.com/ec2/instance-types/
	// - GCP: https://cloud.google.com/compute/docs/machine-resource
	InstanceType string

	// AWS specific configuration
	AWSConfig *AWSConfig

	// GCP Specific configuration
	GCPConfig *GCPConfig
}

type AWSConfig struct {
	// AWS profile to use for the node
	AWSProfile string

	// AWS KeyPair to use for the node
	AWSKeyPair string

	// AWS volume size in GB
	AWSVolumeSize int

	// AWS volume type
	AWSVolumeType string

	// AWS volume IOPS
	AWSVolumeIOPS int

	// AWS volume throughput
	AWSVolumeThroughput int

	// AWS security group to use for the node
	AWSSecurityGroupID string
}

type GCPConfig struct {
	// GCP project to use for the node
	GCPProject string

	// GCP credentials to use for the node
	GCPCredentials string

	// GCP network label to use for the node
	GCPNetwork string

	// GCP zone to use for the node
	GCPZone string

	// GCP Volume size in GB
	GCPVolumeSize int

	// GCP SSH Public Key
	GCPSSHKey string
}

func GetDefaultCloudParams(ctx context.Context, cloud SupportedCloud) (*CloudParams, error) {
	// make sure that CloudParams is initialized with default values
	switch cloud {
	case AWSCloud:
		cp := &CloudParams{
			AWSConfig: &AWSConfig{
				AWSProfile:          "default",
				AWSVolumeSize:       1000,
				AWSVolumeThroughput: 500,
				AWSVolumeIOPS:       1000,
				AWSVolumeType:       "gp3",
			},
			Region:       "us-east-1",
			InstanceType: constants.AWSDefaultInstanceType,
		}
		awsSvc, err := awsAPI.NewAwsCloud(ctx, cp.AWSConfig.AWSProfile, cp.Region)
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
		cp.ImageID = imageID
		return cp, nil
	case GCPCloud:
		projectName, err := getDefaultProjectNameFromGCPCredentials(constants.GCPDefaultAuthKeyPath)
		if err != nil {
			return nil, err
		}
		sshKey, err := GetPublicKeyFromSSHKey("")
		if err != nil {
			return nil, err
		}
		cp := &CloudParams{
			GCPConfig: &GCPConfig{
				GCPProject:     projectName,
				GCPCredentials: utils.ExpandHome(constants.GCPDefaultAuthKeyPath),
				GCPVolumeSize:  constants.CloudServerStorageSize,
				GCPNetwork:     "avalanche-tooling-sdk-go-us-east1",
				GCPSSHKey:      sshKey,
				GCPZone:        "us-east1-b",
			},
			Region:       "us-east1",
			InstanceType: constants.GCPDefaultInstanceType,
		}
		gcpSvc, err := gcpAPI.NewGcpCloud(ctx, cp.GCPConfig.GCPProject, cp.GCPConfig.GCPCredentials)
		if err != nil {
			return nil, err
		}
		imageID, err := gcpSvc.GetUbuntuimageID()
		if err != nil {
			return nil, err
		}
		cp.ImageID = imageID
		return cp, nil
	default:
		return nil, fmt.Errorf("unsupported cloud")
	}
}

// Validate checks that the CloudParams are valid for deployment
func (cp *CloudParams) Validate() error {
	// common checks
	if cp.Region == "" {
		return fmt.Errorf("region is required")
	}
	if cp.ImageID == "" {
		return fmt.Errorf("image is required")
	}
	if cp.InstanceType == "" {
		return fmt.Errorf("instance type is required")
	}
	switch cp.Cloud() {
	case AWSCloud:
		if cp.AWSConfig == nil {
			return fmt.Errorf("AWS config needs to be set")
		}
		if cp.AWSConfig.AWSProfile == "" {
			return fmt.Errorf("AWS profile is required")
		}
		if cp.AWSConfig.AWSSecurityGroupID == "" {
			return fmt.Errorf("AWS security group is required")
		}
		if cp.AWSConfig.AWSVolumeSize < 0 {
			return fmt.Errorf("AWS volume size must be positive")
		}
		if cp.AWSConfig.AWSVolumeType == "" {
			return fmt.Errorf("AWS volume type is required")
		}
		if cp.AWSConfig.AWSVolumeIOPS < 0 {
			return fmt.Errorf("AWS volume IOPS must be positive")
		}
		if cp.AWSConfig.AWSVolumeThroughput < 0 {
			return fmt.Errorf("AWS volume throughput must be positive")
		}
		if cp.AWSConfig.AWSKeyPair == "" {
			return fmt.Errorf("AWS key pair is required")
		}
	case GCPCloud:
		if cp.GCPConfig == nil {
			return fmt.Errorf("AWS config needs to be set")
		}
		if cp.GCPConfig.GCPNetwork == "" {
			return fmt.Errorf("GCP network is required")
		}
		if cp.GCPConfig.GCPProject == "" {
			return fmt.Errorf("GCP project is required")
		}
		if cp.GCPConfig.GCPCredentials == "" {
			return fmt.Errorf("GCP credentials is required")
		}
		if cp.GCPConfig.GCPZone == "" {
			return fmt.Errorf("GCP zone is required")
		}
		if cp.GCPConfig.GCPVolumeSize < 0 {
			return fmt.Errorf("GCP volume size must be positive")
		}
		if !strings.HasPrefix(cp.GCPConfig.GCPZone, cp.Region) {
			return fmt.Errorf("GCP zone must be in the region %s", cp.Region)
		}
		if cp.GCPConfig.GCPSSHKey == "" {
			return fmt.Errorf("GCP SSH key is required")
		}
	default:
		return fmt.Errorf("unsupported cloud")
	}
	return nil
}

// Cloud returns the SupportedCloud for the CloudParams
func (cp *CloudParams) Cloud() SupportedCloud {
	switch {
	case cp.AWSConfig != nil && cp.AWSConfig.AWSProfile != "":
		return AWSCloud
	case cp.GCPConfig != nil && (cp.GCPConfig.GCPProject != "" || cp.GCPConfig.GCPCredentials != ""):
		return GCPCloud
	default:
		return Unknown
	}
}
