// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"log"
	"net"
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
func (e *EchoService) P(args *Args, reply *string) error {
	*reply = args.Message
	return nil
}

// C echoes the message for the C endpoint.
func (e *EchoService) C(args *Args, reply *string) error {
	*reply = args.Message
	return nil
}

// X echoes the message for the X endpoint.
func (e *EchoService) X(args *Args, reply *string) error {
	*reply = args.Message
	return nil
}

func startServer() {
	echo := new(EchoService)
	rpc.Register(echo)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", constants.AvalanchegoAPIPort))
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer listener.Close()

	log.Printf("Serving RPC server on port %d", constants.AvalanchegoAPIPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

// MockNode is a mock implementation of the Node struct.
type MockNode struct{}

// AvalanchegoTCPClient returns a connection to the local RPC server.
func (h *MockNode) AvalanchegoTCPClient() (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf(":%d", constants.AvalanchegoAPIPort))
}

func TestAvalanchegoRPCClient(t *testing.T) {
	// Start the RPC server in a goroutine.
	go startServer()

	// Allow some time for the server to start.
	time.Sleep(time.Second)

	// Create a proxy connection to the local RPC server.
	node := &MockNode{}
	proxy, err := node.AvalanchegoTCPClient()
	if err != nil {
		t.Fatalf("Failed to create proxy connection: %v", err)
	}
	defer proxy.Close()

	fakeNode := &Node{}
	// Test the RPC client with the proxy connection.
	client, err := fakeNode.AvalanchegoRPCClient("", &proxy)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Close()

	// Define the arguments and reply for the RPC call.
	args := &Args{Message: "Hello, World!"}
	var reply string

	// Test the P endpoint.
	err = client.Call("EchoService.P", args, &reply)
	if err != nil {
		t.Fatalf("RPC call to EchoService.P failed: %v", err)
	}
	if reply != args.Message {
		t.Errorf("Expected reply %q, got %q", args.Message, reply)
	}

	// Test the C endpoint.
	err = client.Call("EchoService.C", args, &reply)
	if err != nil {
		t.Fatalf("RPC call to EchoService.C failed: %v", err)
	}
	if reply != args.Message {
		t.Errorf("Expected reply %q, got %q", args.Message, reply)
	}

	// Test the X endpoint.
	err = client.Call("EchoService.X", args, &reply)
	if err != nil {
		t.Fatalf("RPC call to EchoService.X failed: %v", err)
	}
	if reply != args.Message {
		t.Errorf("Expected reply %q, got %q", args.Message, reply)
	}
}
