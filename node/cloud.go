// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

// AWS specific configuration
type AWSSpec struct {
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

	// AWS Elastic IP to use for the node
	ElasticIP string

	// AWS security group to use for the node
	SecurityGroup string

	// AWS Region to use for the node
	Region string

	// AWS AMI id to use for the node
	Image string

	// AWS Instance type to use for the node
	InstanceType string
}

type GCPSpec struct {
	// GCP project to use for the node
	Project string

	// GCP credentials to use for the node
	Credentials string

	// GCP static IP to use for the node
	StaticIP string

	// GCP network label to use for the node
	Network string

	// GCP Image to use for the node
	Image string

	// GCP Instance type to use for the node
	InstanceType string
}
