// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

import (
	"fmt"
	"net"
	"net/rpc"
	"net/url"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// AvalanchegoTCPClient returns the connection to the node.
func (h *Host) AvalanchegoTCPClient() (*net.Conn, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	avalancheGoEndpoint := strings.TrimPrefix(constants.LocalAPIEndpoint, "http://")
	avalancheGoAddr, err := net.ResolveTCPAddr("tcp", avalancheGoEndpoint)
	if err != nil {
		return nil, err
	}
	var proxy net.Conn
	proxy, err = h.Connection.DialTCP("tcp", nil, avalancheGoAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to port forward to %s via %s", h.Connection.RemoteAddr(), "ssh")
	}
	return &proxy, nil
}

// AvalanchegoRPCClient returns the RPC client to the node.
func (h *Host) AvalanchegoRPCClient() (*rpc.Client, error) {
	proxy, err := h.AvalanchegoTCPClient()
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(*proxy), nil
}

// Post sends a POST request to the node at the specified path with the provided body.
func (h *Host) Post(path string, requestBody string) ([]byte, error) {
	if path == "" {
		path = "/ext/info"
	}
	localhost, err := url.Parse(constants.LocalAPIEndpoint)
	if err != nil {
		return nil, err
	}
	requestHeaders := fmt.Sprintf("POST %s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Content-Length: %d\r\n"+
		"Content-Type: application/json\r\n\r\n", path, localhost.Host, len(requestBody))
	httpRequest := requestHeaders + requestBody
	return h.Forward(httpRequest, constants.SSHPOSTTimeout)
}
