package host

import "sync"

// HostResult is a struct that holds the result of a async command executed on a host
type HostResult struct {
	// ID of the host
	NodeID string

	// Value is the result of the command executed on the host
	Value interface{}

	// Err is the error that occurred while executing the command on the host
	Err error
}

// HostResults is a struct that holds the results of multiple async commands executed on multiple hosts
type HostResults struct {
	Results []HostResult
	Lock    sync.Mutex
}

// AddResult adds a result to the HostResults
func (nr *HostResults) AddResult(nodeID string, value interface{}, err error) {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	nr.Results = append(nr.Results, HostResult{
		NodeID: nodeID,
		Value:  value,
		Err:    err,
	})
}

// GetResults returns the results of the HostResults
func (nr *HostResults) GetResults() []HostResult {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	return nr.Results
}

// GetResultMap returns a map of the results of the HostResults with the nodeID as the key
func (nr *HostResults) GetResultMap() map[string]interface{} {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	result := map[string]interface{}{}
	for _, node := range nr.Results {
		result[node.NodeID] = node.Value
	}
	return result
}

// GetErrorMap returns a map of the errors of the HostResults with the nodeID as the key
func (nr *HostResults) Len() int {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	return len(nr.Results)
}

// GetNodeList returns a list of the nodeIDs of the HostResults
func (nr *HostResults) GetNodeList() []string {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	nodes := []string{}
	for _, node := range nr.Results {
		nodes = append(nodes, node.NodeID)
	}
	return nodes
}

// GetErrorMap returns a map of the errors of the HostResults with the nodeID as the key
func (nr *HostResults) GetErrorHostMap() map[string]error {
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

// HasNodeIDWithError checks if a node with the given nodeID has an error
func (nr *HostResults) HasNodeIDWithError(nodeID string) bool {
	nr.Lock.Lock()
	defer nr.Lock.Unlock()
	for _, node := range nr.Results {
		if node.NodeID == nodeID && node.Err != nil {
			return true
		}
	}
	return false
}

// HasErrors returns true if the HostResults has any errors
func (nr *HostResults) HasErrors() bool {
	return len(nr.GetErrorHostMap()) > 0
}

// GetErrorHosts returns a list of the nodeIDs of the HostResults that have errors
func (nr *HostResults) GetErrorHosts() []string {
	var nodes []string
	for _, node := range nr.Results {
		if node.Err != nil {
			nodes = append(nodes, node.NodeID)
		}
	}
	return nodes
}
