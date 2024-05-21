// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

// CloudConfig contains the configuration for deploying a node in the cloud
type CloudConfig struct {
	// Cloud region to deploy the node in
	Region string

	// Cloud image to use for the node.
	// For AWS it's the AMI ID, for GCP it's the image name
	Image string

	// Cloud key pair to use for the node
	KeyPair string

	// Cloud security group to use for the node.
	SecurityGroup string

	// Cloud static IP assigned to the node
	// It's empty if the node has ephemeral IP
	ElasticIP string

	// Cloud instance type to use for the node
	InstanceType string

	// Cloud specific configuration
	CloudSpec interface{}
}

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
}

type GCPSpec struct {
	// GCP project to use for the node
	Project string

	// GCP credentials to use for the node
	Credentials string
}
