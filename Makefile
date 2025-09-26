# Makefile for Avalanche Tooling SDK Go

.PHONY: proto generate build test clean

# Generate gRPC code from proto files
proto:
	@echo "Generating gRPC code..."
	@mkdir -p api/generated
	protoc --go_out=api/generated --go_opt=paths=source_relative \
		--go-grpc_out=api/generated --go-grpc_opt=paths=source_relative \
		api/proto/*.proto
	@echo "gRPC code generated successfully"

# Generate all code
generate: proto

# Build the project
build: generate
	go build ./...

# Run tests
test:
	go test ./...

# Clean generated files
clean:
	rm -rf api/generated

# Install required tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Generate with HTTP gateway
proto-gateway:
	@echo "Generating gRPC code with HTTP gateway..."
	@mkdir -p api/generated
	protoc -I. -I./third_party/googleapis \
		--go_out=api/generated --go_opt=paths=source_relative \
		--go-grpc_out=api/generated --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=api/generated --grpc-gateway_opt=paths=source_relative \
		api/proto/*.proto
	@echo "gRPC code with HTTP gateway generated successfully"

# Server API examples
.PHONY: run-server run-client run-simple-client run-complete-example run-http-server run-http-client run-http-demo

# Run the gRPC server
run-server:
	@echo "Starting gRPC server on port 8080..."
	go run test_server.go

# Run the HTTP client (gRPC)
run-client:
	@echo "Running gRPC client..."
	go run test_client.go

# Run the combined server (gRPC + HTTP)
run-http-server:
	@echo "Starting combined server (gRPC on :8080, HTTP on :8081)..."
	go run http_server_example.go

# Run the HTTP client
run-http-client:
	@echo "Running HTTP client..."
	go run http_client_example.go

# Run both HTTP server and client (requires two terminals)
run-http-demo: run-http-server run-http-client

# Legacy examples (if they exist)
run-simple-client:
	@echo "Running simple client example..."
	go run examples/simple_client_example.go

# Run the complete client example
run-complete-example:
	@echo "Running complete client example..."
	go run examples/complete_server_example.go

# Run both server and client (requires two terminals)
run-demo: run-server run-client
