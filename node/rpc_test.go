// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

// Args represents the arguments passed to the RPC methods.
type Args struct {
	Message string
}

// EchoService provides methods to echo messages.
type EchoService struct{}

// P echoes the message for the P endpoint.
func (e *EchoService) Echo(args *Args, reply *string) error {
	*reply = args.Message
	return nil
}

func startMockServer(ctx context.Context, path string) {
	echo := new(EchoService)
	server := rpc.NewServer()
	server.Register(echo)
	mux := http.NewServeMux()
	mux.Handle(path, server)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", constants.AvalanchegoAPIPort))
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer listener.Close()

	srv := &http.Server{
		Handler: mux,
	}
	log.Printf("Serving RPC server on port %d", constants.AvalanchegoAPIPort)

	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		srv.Shutdown(context.Background())
	}()

	log.Printf("Serving RPC server on port %d", constants.AvalanchegoAPIPort)
	if err := srv.Serve(listener); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

// MockNode is a mock implementation of the Node struct.
type MockNode struct{}

// AvalanchegoTCPClient returns a connection to the local RPC server.
func (h *MockNode) AvalanchegoTCPClient() (net.Conn, error) {
	return net.Dial("tcp", "127.0.0.1:9650")
}

// AvalanchegoRPCClient returns the RPC client to the node.
func (h *MockNode) AvalanchegoRPCClient(chainID string, proxy *net.Conn) (*rpc.Client, error) {
	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(*proxy))
	return client, nil
}

// TestAvalanchegoRPCClient tests the RPC client using a proxy connection.
func TestAvalanchegoRPCClient(t *testing.T) {
	// Start the RPC server in a goroutine.
	ctx, cancelRoot := context.WithCancel(context.Background())
	go startMockServer(ctx, "/")

	// Allow some time for the server to start.
	time.Sleep(time.Second)

	// Create a proxy connection to the local RPC server.
	node := &MockNode{}
	proxy, err := node.AvalanchegoTCPClient()
	if err != nil {
		t.Fatalf("Failed to create proxy connection: %v", err)
	}
	defer proxy.Close()

	// Log to check if proxy connection is working.
	t.Logf("Proxy connection established to %s", proxy.RemoteAddr().String())

	// Test the RPC client with the proxy connection.
	client, err := node.AvalanchegoRPCClient("", &proxy)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Close()

	// Define the arguments and reply for the RPC call.
	args := &Args{Message: "Hello, World!"}
	var reply string

	// Test the P endpoint.
	err = client.Call("EchoService.Echo", args, &reply)
	if err != nil {
		t.Fatalf("RPC call to EchoService.ECho failed: %v", err)
	}
	if reply != args.Message {
		t.Errorf("Expected reply %q, got %q", args.Message, reply)
	}
	cancelRoot()

}
