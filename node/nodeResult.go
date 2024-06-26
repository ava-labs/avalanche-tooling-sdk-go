// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package node

import (
	"fmt"
	"sync"
)

// NodeResult is a struct that holds the result of a async command executed on a host
type NodeResult struct {
	// ID of the host
	NodeID string

	// Value is the result of the command executed on the host
	Value interface{}

	// Err is the error that occurred while executing the command on the host
	Err error
}

// NodeResults is a struct that holds the results of multiple async commands executed on multiple hosts
type NodeResults struct {
	Results []NodeResult
	Lock    sync.Mutex
}

// AddResult adds a new NodeResult to the NodeResults struct.
//
// Parameters:
// - nodeID: the ID of the host.
// - value: the result of the command executed on the host.
// - err: the error that occurred while executing the command on the host.
func (nr *NodeResults) AddResult(nodeID string, value interface{}, err error) {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	nr.Results = append(nr.Results, NodeResult{
		NodeID: nodeID,
		Value:  value,
		Err:    err,
	})
}

// GetResults returns the results of the NodeResults
//
// No parameters.
// Returns:
// - []NodeResult: the results of the NodeResults.
func (nr *NodeResults) GetResults() []NodeResult {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	return nr.Results
}

// GetResultMap returns a map of the results of the NodeResults with the nodeID as the key.
//
// It acquires the lock on the NodeResults and iterates over the Results slice.
// For each NodeResult, it adds the NodeID as the key and the Value as the value to the result map.
// Finally, it releases the lock and returns the result map.
//
// Returns:
// - map[string]interface{}: A map with the nodeIDs as keys and the corresponding values as values.
func (nr *NodeResults) GetResultMap() map[string]interface{} {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	result := map[string]interface{}{}
	for _, node := range nr.Results {
		result[node.NodeID] = node.Value
	}
	return result
}

// Len returns the number of results in the NodeResults.
//
// It acquires the lock on the NodeResults and returns the length of the Results slice.
// The lock is released before the function returns.
//
// Returns:
// - int: the number of results in the NodeResults.
func (nr *NodeResults) Len() int {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	return len(nr.Results)
}

// GetNodeList returns a list of the nodeIDs of the NodeResults.
//
// No parameters.
// Returns a slice of strings.
func (nr *NodeResults) GetNodeList() []string {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	nodes := []string{}
	for _, node := range nr.Results {
		nodes = append(nodes, node.NodeID)
	}
	return nodes
}

// GetErrorHostMap returns a map of the errors of the NodeResults with the nodeID as the key.
//
// It acquires the lock on the NodeResults and iterates over the Results slice.
// For each NodeResult, if the Err field is not nil, it adds the NodeID as the key and the error as the value to the hostErrors map.
// Finally, it releases the lock and returns the hostErrors map.
//
// Returns:
// - map[string]error: A map with the nodeIDs as keys and the corresponding errors as values.
func (nr *NodeResults) GetErrorHostMap() map[string]error {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	hostErrors := make(map[string]error)
	for _, node := range nr.Results {
		if node.Err != nil {
			hostErrors[node.NodeID] = node.Err
		}
	}
	return hostErrors
}

// HasNodeIDWithError checks if a node with the given nodeID has an error.
//
// Parameters:
// - nodeID: the ID of the node to check.
//
// Return:
// - bool: true if a node with the given nodeID has an error, false otherwise.
func (nr *NodeResults) HasNodeIDWithError(nodeID string) bool {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	for _, node := range nr.Results {
		if node.NodeID == nodeID && node.Err != nil {
			return true
		}
	}
	return false
}

// HasErrors returns true if the NodeResults has any errors.
//
// It checks the length of the error host map obtained from the GetErrorHostMap()
// method of the NodeResults struct. If the length is greater than 0, it means
// that there are errors present, and the function returns true. Otherwise, it
// returns false.
func (nr *NodeResults) HasErrors() bool {
	return len(nr.GetErrorHostMap()) > 0
}

// GetErrorHosts returns a list of the nodeIDs of the NodeResults that have errors.
//
// No parameters.
// Returns a slice of strings.
func (nr *NodeResults) GetErrorHosts() []string {
	var nodes []string
	for _, node := range nr.Results {
		if node.Err != nil {
			nodes = append(nodes, node.NodeID)
		}
	}
	return nodes
}

// SumError collects and returns the errors with nodeIds if there are errors in the NodeResults.
//
// Returns an error type.
func (nr *NodeResults) Error() error {
	if nr.HasErrors() {
		// if there are errors, collect and return them with nodeIds
		hostErrorMap := nr.GetErrorHostMap()
		errStr := ""
		for nodeID, err := range hostErrorMap {
			errStr += fmt.Sprintf("NodeID: %s, Error: %s\n", nodeID, err)
		}
		return fmt.Errorf(errStr)
	} else {
		return nil
	}
}
