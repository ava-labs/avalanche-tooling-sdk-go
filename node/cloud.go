// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

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
