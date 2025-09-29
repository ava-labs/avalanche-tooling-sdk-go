# Makefile for Avalanche Tooling SDK Go

.PHONY: proto generate build test clean run-http-server install-tools

# Generate gRPC code from proto files
proto:
	@echo "Generating gRPC code..."
	@mkdir -p api/generated
	protoc -I. -I./third_party/googleapis \
		--go_out=api/generated --go_opt=paths=source_relative \
		--go-grpc_out=api/generated --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=api/generated --grpc-gateway_opt=paths=source_relative \
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

# Run the HTTP server (gRPC + HTTP)
run-http-server:
	@echo "Starting combined server (gRPC on :8080, HTTP on :8081)..."
	go run http_server_example.go
