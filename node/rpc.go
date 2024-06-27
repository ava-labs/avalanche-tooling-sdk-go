// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

type httpTransport struct {
	Transport http.RoundTripper
	Scope     string
	conn      net.Conn
}

func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.HasPrefix(req.URL.Path, "/ext/"+t.Scope) {
		return nil, http.ErrNotSupported
	}
	return t.Transport.RoundTrip(req)
}

func (t *httpTransport) Read(p []byte) (int, error) {
	return t.conn.Read(p)
}

func (t *httpTransport) Write(p []byte) (int, error) {
	return t.conn.Write(p)
}

func (t *httpTransport) Close() error {
	return t.conn.Close()
}

// AvalanchegoRPCClient returns the RPC client to the node.
func (h *Node) AvalanchegoRPCClient(chainID string, proxy *net.Conn) (*rpc.Client, error) {
	var err error
	// proxy passed should be non unless in testing
	if proxy == nil {
		proxy, err = h.AvalanchegoTCPClient()
		if err != nil {
			return nil, err
		}
	}
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return *proxy, nil
		},
	}

	if chainID != "" {
		client := rpc.NewClient(&httpTransport{
			Transport: transport,
			Scope:     chainID,
			conn:      *proxy,
		})
		return client, nil
	} else {
		return rpc.NewClient(&httpTransport{
			Transport: transport,
			conn:      *proxy,
		}), nil
	}
}
