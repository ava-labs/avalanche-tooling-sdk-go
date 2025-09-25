# Avalanche Tooling SDK - Server API Usage Guide

## Overview

The Avalanche Tooling SDK provides a gRPC-based server API that allows you to manage Avalanche wallets and accounts remotely. The server exposes two main services:

- **WalletService**: Wallet management and transaction operations
- **AccountService**: Account-specific operations

## Quick Start

### 1. Start the Server

```bash
# Run the server on port 8080
go run examples/grpc_server_example.go
```

Or use the test server:
```bash
go run test_server.go
```

### 2. Connect as a Client

```bash
# Run the client example
go run examples/grpc_client_example.go
```

## Available Services

### WalletService

#### CreateAccount
Creates a new account with addresses for Fuji, Mainnet, and Ethereum networks.

**Request:**
```protobuf
message CreateAccountRequest {
  // Empty for now, could add account options later
}
```

**Response:**
```protobuf
message CreateAccountResponse {
  string fuji_avax_address = 1;    // Fuji testnet address
  string avax_address = 2;         // Mainnet address
  string eth_address = 3;          // Ethereum address
}
```

#### GetAccount
Retrieve account information by address.

**Request:**
```protobuf
message GetAccountRequest {
  string address = 1;  // Account address to retrieve
}
```

**Response:**
```protobuf
message GetAccountResponse {
  string address = 1;        // Account address
  repeated string policies = 2;  // Account policies
  string created_at = 3;     // Creation timestamp in ISO 8601 format (e.g., "2025-03-25T12:00:00Z")
  string updated_at = 4;     // Last modification timestamp in ISO 8601 format (e.g., "2025-03-25T12:00:00Z")
  string key_type = 5;       // Key type (e.g., "secp-ava", "secp-ava-test", "secp")
}
```

#### Other Methods (Currently Unimplemented)
- `ListAccounts` - List all accounts
- `ImportAccount` - Import existing account
- `BuildTransaction` - Build transactions
- `SignTransaction` - Sign transactions
- `SendTransaction` - Send transactions
- `GetChainClients` - Get chain client endpoints
- `SetChainClients` - Set chain client endpoints
- `Close` - Cleanup resources

### AccountService

#### GetPChainAddress (Unimplemented)
Get P-Chain address for a specific account and network.

#### GetKeychain (Unimplemented)
Get keychain information for a specific account.

## Configuration

### Session Management
The server uses Cubist's session management for key derivation. You need to have a valid session file at:
```
/Users/raymondsukanto/Desktop/management-session.json
```

### Account Storage
Accounts are stored in a JSON file at:
```
data/accounts.json
```

## Example Usage

### Go Client Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Create API wallet that connects to gRPC server
    apiWallet, err := wallet.NewAPIWallet("localhost:8080")
    if err != nil {
        log.Fatalf("Failed to create API wallet: %v", err)
    }
    defer apiWallet.Close(ctx)

    // Create a new account
    account, err := apiWallet.CreateAccount(ctx)
    if err != nil {
        log.Fatalf("Failed to create account: %v", err)
    }

    fmt.Printf("Created account with addresses: %v\n", account.Addresses())
}
```

### Using gRPC Directly

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create client
    client := proto.NewWalletServiceClient(conn)

    // Create account
    resp, err := client.CreateAccount(context.Background(), &proto.CreateAccountRequest{})
    if err != nil {
        log.Fatalf("Failed to create account: %v", err)
    }

    log.Printf("Fuji Address: %s", resp.FujiAvaxAddress)
    log.Printf("Mainnet Address: %s", resp.AvaxAddress)
    log.Printf("ETH Address: %s", resp.EthAddress)
}
```

## Error Handling

The server returns gRPC status codes:
- `OK` - Success
- `Internal` - Server error
- `Unimplemented` - Method not yet implemented
- `InvalidArgument` - Invalid request parameters

## Development Notes

### Current Limitations
1. Most methods are not yet implemented (return `Unimplemented` status)
2. Only `CreateAccount` is fully functional
3. Session file path is hardcoded
4. Account storage uses local JSON files

### Future Enhancements
1. Implement remaining service methods
2. Add proper error handling and validation
3. Support for multiple session management backends
4. Database storage for accounts
5. Authentication and authorization
6. Configuration management

## Troubleshooting

### Common Issues

1. **Session file not found**
   - Ensure the session file exists at the specified path
   - Check file permissions

2. **Port already in use**
   - Change the port in the server configuration
   - Kill existing processes using the port

3. **Account creation fails**
   - Verify Cubist session is valid
   - Check network connectivity
   - Ensure proper permissions for data directory

### Debugging

Enable gRPC reflection for debugging:
```go
reflection.Register(grpcServer)
```

Use tools like `grpcurl` to test the API:
```bash
# List services
grpcurl -plaintext localhost:8080 list

# Call CreateAccount
grpcurl -plaintext localhost:8080 avalanche.wallet.v1.WalletService/CreateAccount
```
