// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package installer

import (
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/node"
)

type NodeInstaller struct {
	Node *node.Node
}

func NewHostInstaller(node *node.Node) *NodeInstaller {
	return &NodeInstaller{Node: node}
}

func (i *NodeInstaller) GetArch() (string, string) {
	goArhBytes, err := i.Node.Command(nil, constants.SSHScriptTimeout, "dpkg --print-architecture")
	if err != nil {
		return "", ""
	}
	goOSBytes, err := i.Node.Command(nil, constants.SSHScriptTimeout, "uname -s")
	if err != nil {
		return "", ""
	}
	return strings.TrimSpace(string(goArhBytes)), strings.TrimSpace(strings.ToLower(string(goOSBytes)))
}
