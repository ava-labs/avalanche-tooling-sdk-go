# Avalanche Tooling SDK - Examples

This directory contains examples demonstrating how to use the Avalanche Tooling SDK server API.

## Quick Start

1. **Start the server:**
   ```bash
   make run-server
   # or
   go run grpc_server_example.go
   ```

2. **Run a client example:**
   ```bash
   make run-simple-client
   # or
   go run simple_client_example.go
   ```

## Examples

### grpc_server_example.go
A simple gRPC server that implements the Avalanche wallet API. This server:
- Runs on port 8080
- Implements the WalletService and AccountService
- Currently only supports account creation
- Uses Cubist for key management

### grpc_client_example.go
A comprehensive client example that demonstrates:
- Connecting to the gRPC server
- Creating accounts
- Building transactions
- Using the wallet API wrapper

### simple_client_example.go
A simplified client example that shows:
- Basic connection to the server
- Account creation
- Error handling for unimplemented methods

### complete_server_example.go
A detailed example that demonstrates:
- Direct gRPC client usage
- All available API methods
- Error handling and status checking
- Both WalletService and AccountService calls

## Prerequisites

1. **Session File**: The server requires a Cubist session file at:
   ```
   /Users/raymondsukanto/Desktop/management-session.json
   ```

2. **Go Dependencies**: Install required dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Using Make Commands

```bash
# Start server
make run-server

# Run simple client
make run-simple-client

# Run complete example
make run-complete-example

# Run quick start script
./quickstart.sh
```

### Manual Execution

1. **Terminal 1 - Start Server:**
   ```bash
   go run examples/grpc_server_example.go
   ```

2. **Terminal 2 - Run Client:**
   ```bash
   go run examples/simple_client_example.go
   ```

## API Status

| Method | Status | Description |
|--------|--------|-------------|
| CreateAccount | ✅ Implemented | Creates new accounts with Fuji, Mainnet, and ETH addresses |
| GetAccount | ❌ Unimplemented | Retrieves account by address |
| ListAccounts | ❌ Unimplemented | Lists all accounts |
| ImportAccount | ❌ Unimplemented | Imports existing account |
| BuildTransaction | ❌ Unimplemented | Builds transactions |
| SignTransaction | ❌ Unimplemented | Signs transactions |
| SendTransaction | ❌ Unimplemented | Sends transactions |
| GetChainClients | ❌ Unimplemented | Gets chain client endpoints |
| SetChainClients | ❌ Unimplemented | Sets chain client endpoints |
| GetPChainAddress | ❌ Unimplemented | Gets P-Chain address |
| GetKeychain | ❌ Unimplemented | Gets keychain information |

## Troubleshooting

### Common Issues

1. **"Session file not found"**
   - Ensure the session file exists at the specified path
   - Check file permissions

2. **"Connection refused"**
   - Make sure the server is running
   - Check if port 8080 is available

3. **"Method not implemented"**
   - This is expected for most methods
   - Only CreateAccount is currently implemented

### Debugging

Use gRPC reflection for debugging:
```bash
# List available services
grpcurl -plaintext localhost:8080 list

# Call CreateAccount
grpcurl -plaintext localhost:8080 avalanche.wallet.v1.WalletService/CreateAccount
```

## Development

To add new functionality:

1. Update the proto files in `api/proto/`
2. Regenerate gRPC code: `make proto`
3. Implement the server methods in `api/server/`
4. Update client examples as needed

For more information, see the main [SERVER_API_GUIDE.md](../SERVER_API_GUIDE.md).
