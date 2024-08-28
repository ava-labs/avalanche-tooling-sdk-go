// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"strings"
	"time"
)

func (h *Node) GetContainerLogs(containerName string, tailLines uint, timeout time.Duration) ([]string, error) {
	if containerName == "" {
		return nil, fmt.Errorf("container name cannot be empty")
	}
	tailLinesString := "all"
	if tailLines > 0 {
		tailLinesString = fmt.Sprintf("%d", tailLines)
	}
	output, err := h.Commandf(nil, timeout, "docker logs --tail %s %s", tailLinesString, containerName)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(output), "\n"), nil
}

func (h *Node) GetAvalanchegoLogs(tailLines uint, timeout time.Duration) ([]string, error) {
	return h.GetContainerLogs("avalanchego", tailLines, timeout)
}

func (h *Node) GetAWMRelayerLogs(tailLines uint, timeout time.Duration) ([]string, error) {
	return h.GetContainerLogs("awm-relayer", tailLines, timeout)
}
