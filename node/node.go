// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

// SSHConfig contains the configuration for connecting to a node over SSH
type SSHConfig struct {
	// Username to use when connecting to the nodeßß
	User string

	// Path to the private key to use when connecting to the node
	// If this is empty, the SSH agent will be used
	KeyPath string

	// Parameters to pass to the ssh command.
	// See man ssh_config(5) for more information
	// By defalult it's StrictHostKeyChecking=no
	Params map[string]string // additional parameters to pass to the ssh command
}

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

	// Cloud type to deploy the node in
	Cloud SupportedCloud

	// Cloud static IP assigned to the node
	// It's empty if the node has ephemeral IP
	ElasticIP string

	// Roles of the node
	Roles []SupportedRole
}

type Node struct {
	// ID of the node
	ID string

	// IP address of the node
	IP string

	// SSH configuration for the node
	SSHConfig SSHConfig

	// Cloud configuration for the node
	CloudConfig CloudConfig
}
