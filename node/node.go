// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

// SSHConfig contains the configuration for connecting to a node over SSH
type SSHConfig struct {
	// Username to use when connecting to the node
	user string

	// Path to the private key to use when connecting to the node
	// If this is empty, the SSH agent will be used
	KeyPath string

	// Parameters to pass to the ssh command.
	// See man ssh_config(5) for more information
	// By defalult it's StrictHostKeyChecking=no
	Params map[string]string // additional parameters to pass to the ssh command
}

type Node struct {
	// ID of the node
	ID string

	// IP address of the node
	IP string

	// SSH configuration for the node
	SSHConfig SSHConfig

	// Cloud configuration for the node
	Cloud SupportedCloud

	// CloudConfig is the cloud specific configuration for the node
	CloudConfig interface{}

	// Roles of the node
	Roles []SupportedRole
}
