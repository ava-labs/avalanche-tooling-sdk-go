// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"strings"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// AvalanchegoTCPClient returns the connection to the node.
func (h *Node) AvalanchegoTCPClient() (*net.Conn, error) {
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
	proxy, err = h.connection.DialTCP("tcp", nil, avalancheGoAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to port forward to %s via %s", h.connection.RemoteAddr(), "ssh")
	}
	return &proxy, nil
}

type httpTransport struct {
	Transport http.RoundTripper
	Scope     string
}

func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.HasPrefix(req.URL.Path, "/ext/"+t.Scope) {
		return nil, http.ErrNotSupported
	}
	return t.Transport.RoundTrip(req)
}

// AvalanchegoRPCClient returns the RPC client to the node.
func (h *Node) AvalanchegoRPCClient(chainID string) (*rpc.Client, error) {
	proxy, err := h.AvalanchegoTCPClient()
	if err != nil {
		return nil, err
	}
	if chainID != "" {
		client := rpc.NewClient(&httpTransport{
			Transport: proxy.Transport,
			Scope:     scope,
		})
		return client, nil
	} else {
		return rpc.NewClient(*proxy), nil
	}
}

type scopeFilteringTransport struct {
	Transport http.RoundTripper
	Scope     string
}

func (t *scopeFilteringTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.HasPrefix(req.URL.Path, "/ext/"+t.Scope) {
		return nil, http.ErrNotSupported
	}
	return t.Transport.RoundTrip(req)
}

// Post sends a POST request to the node at the specified path with the provided body.
func (h *Node) Post(path string, requestBody string) ([]byte, error) {
	if path == "" {
		path = "/ext/info"
	}
	localhost, err := url.Parse(constants.LocalAPIEndpoint)
	if err != nil {
		return nil, err
	}
	requestHeaders := fmt.Sprintf("POST %s HTTP/1.1\r\n"+
		"Node: %s\r\n"+
		"Content-Length: %d\r\n"+
		"Content-Type: application/json\r\n\r\n", path, localhost.Host, len(requestBody))
	httpRequest := requestHeaders + requestBody
	return h.Forward(httpRequest, constants.SSHPOSTTimeout)
}

// WaitForPort waits for the SSH port to become available on the node.
func (h *Node) WaitForPort(port uint, timeout time.Duration) error {
	if port == 0 {
		port = constants.SSHTCPPort
	}
	start := time.Now()
	deadline := start.Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: SSH port %d on node %s is not available after %vs", port, h.IP, timeout.Seconds())
		}
		if _, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", h.IP, port), time.Second); err == nil {
			return nil
		}
		time.Sleep(constants.SSHSleepBetweenChecks)
	}
}
