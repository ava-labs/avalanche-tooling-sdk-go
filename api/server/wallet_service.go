// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	fujiKeyType    = "secp-ava-test"
	mainnetKeyType = "secp-ava"
	ethKeytype     = "secp"
)

// KeyInfo represents a key from the CLI response
type KeyInfo struct {
	Created      int64         `json:"created"`
	LastModified int64         `json:"last_modified"`
	Version      int           `json:"version"`
	Enabled      bool          `json:"enabled"`
	KeyID        string        `json:"key_id"`
	KeyType      string        `json:"key_type"`
	MaterialID   string        `json:"material_id"`
	Owner        string        `json:"owner"`
	Policy       []interface{} `json:"policy"`
	PublicKey    string        `json:"public_key"`
	Purpose      string        `json:"purpose"`
}

// CreateKeyResponse represents the response from the create key command
type CreateKeyResponse struct {
	Keys []KeyInfo `json:"keys"`
}

// WalletServer implements the gRPC WalletService
type WalletServer struct {
	proto.UnimplementedWalletServiceServer
	*primary.Wallet
	accounts                 []account.Account
	signerCliBinaryPath      string
	signerSessionManagerPath string
}

// ImportKey imports a key using the CLI binary
func (s *WalletServer) ImportKey(ctx context.Context, keyType, keyFilePath string) ([]byte, error) {
	// Validate inputs
	if keyType == "" {
		return nil, fmt.Errorf("key type cannot be empty")
	}
	if keyFilePath == "" {
		return nil, fmt.Errorf("key file path cannot be empty")
	}
	if !utils.FileExists(keyFilePath) {
		return nil, fmt.Errorf("key file does not exist: %s", keyFilePath)
	}

	// Build CLI command
	args := []string{
		"key", "import",
		"--key-type", keyType,
		"raw-key",
		"--key-file", keyFilePath,
	}

	// Execute command with session management
	output, err := s.executeCLICommandWithSession(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to import key via CLI: %w", err)
	}
	return output, nil
}

func (s *WalletServer) CreateKeyMnemonic(ctx context.Context) (string, error) {
	// Build CLI command
	args := []string{
		"key", "create",
		"--key-type", "mnemonic",
	}

	// Execute command with session management
	output, err := s.executeCLICommandWithSession(ctx, args)
	if err != nil {
		fmt.Printf("obtained output %s \n", output)
		return "", fmt.Errorf("failed to create key via CLI: %w", err)
	}

	// Parse the JSON response
	var createKeyResp CreateKeyResponse
	if err := json.Unmarshal(output, &createKeyResp); err != nil {
		return "", fmt.Errorf("failed to parse CLI response: %w", err)
	}

	// Check if we have at least one key
	if len(createKeyResp.Keys) == 0 {
		return "", fmt.Errorf("no keys found in CLI response")
	}

	// Return the material_id of the first key
	return createKeyResp.Keys[0].MaterialID, nil
}
func (s *WalletServer) DeriveKeyFromMnemonic(ctx context.Context, keyType string, mnemonicId string) ([]byte, error) {
	//./cs key derive --key-type secp-ava-test --mnemonic-id 0x77d33975e4a52019e869091ef343083d8f11542b262bdfb7fa6d7597ebc97f4c --derivation-path "m/44'/9000'/0'/0/0"
	//./cs key derive \
	//>   --mnemonic-id 0x77d33975e4a52019e869091ef343083d8f11542b262bdfb7fa6d7597ebc97f4c \
	//>   --key-type secp \
	//>   -d "m/44'/60'/0'/0/0" \
	derivationPath := ""
	// Validate inputs
	if keyType == "" {
		return nil, fmt.Errorf("key type cannot be empty")
	}
	if keyType == fujiKeyType || keyType == mainnetKeyType {
		derivationPath = "m/44'/9000'/0'/0/0"
	} else if keyType == ethKeytype {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	// Build CLI command
	args := []string{
		"key", "derive",
		"--key-type", keyType,
		"--mnemonic-id", mnemonicId,
		"-d", derivationPath,
	}

	// Execute command with session management
	output, err := s.executeCLICommandWithSession(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to import key via CLI: %w", err)
	}
	return output, nil
}

// ImportKeyRaw imports a raw key using the CLI binary (like your example)
func (s *WalletServer) ImportKeyRaw(ctx context.Context, keyType string, keyFilePath string) ([]byte, error) {
	return s.ImportKey(ctx, keyType, keyFilePath)
}

// executeCLICommandWithSession executes a CLI command using session file
func (s *WalletServer) executeCLICommandWithSession(ctx context.Context, args []string) ([]byte, error) {
	//TODO: refresh session if it expires

	// Print the command being executed
	fmt.Printf("Executing CLI command: %s %s\n", s.signerCliBinaryPath, strings.Join(args, " "))

	// Execute the CLI command
	cmd := exec.CommandContext(ctx, s.signerCliBinaryPath, args...)
	output, err := cmd.Output()

	fmt.Printf("CLI command output initial: %s\n", string(output))

	// Print the output
	if err != nil {
		fmt.Printf("CLI command failed with error: %v\n", err)
	} else {
		fmt.Printf("CLI command output: %s\n", string(output))
	}

	return output, err
}

// getCLIBinaryPath returns the path to the CLI binary
func getCLIBinaryPath() (string, error) {
	// Check if CLI binary exists in cubesigner directory
	cliPath := filepath.Join("cubesigner", "cs")
	if utils.FileExists(cliPath) {
		return cliPath, nil
	}
	return "", fmt.Errorf("cubesigner CLI binary not found")
}

func getCubeSignerSessionPath() (string, error) {
	// Check if CLI binary exists in cubesigner directory
	sessionPath := filepath.Join("cubesigner", "newSession.json")
	if utils.FileExists(sessionPath) {
		return sessionPath, nil
	}
	return "", fmt.Errorf("cubesigner session manager not found")
}

// NewWalletServer creates a new WalletServer
func NewWalletServer() (*WalletServer, error) {
	// Get CLI binary path
	cliPath, err := getCLIBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cubesigner CLI binary path: %w", err)
	}

	// Get session file path
	sessionPath, err := getCubeSignerSessionPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cubesigner session manager path: %w", err)
	}

	return &WalletServer{
		accounts:                 []account.Account{},
		signerCliBinaryPath:      cliPath,
		signerSessionManagerPath: sessionPath,
	}, nil
}

// CreateAccount creates a new account
func (s *WalletServer) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.CreateAccountResponse, error) {
	// Define all key types to iterate through
	keyTypes := []string{
		fujiKeyType,    // "secp-ava-test"
		mainnetKeyType, // "secp-ava"
		ethKeytype,     // "secp"
	}

	var derivedKeys [][]byte
	var errors []error

	// Generate a new private key
	//k, err := key.NewSoft()
	//if err != nil {
	//	return nil, status.Errorf(codes.Internal, "failed to generate private key: %v", err)
	//}
	//privateKeyHex := k.PrivKeyHex()
	//fmt.Printf("obtained key hex %s \n", privateKeyHex)
	// Create a temporary directory in the repo for key files
	//tempDir := "temp_keys"
	//err = os.MkdirAll(tempDir, 0755)
	//if err != nil {
	//	return nil, status.Errorf(codes.Internal, "failed to create temp directory: %v", err)
	//}
	//defer os.RemoveAll(tempDir) // Clean up the entire temp directory

	//tempFilePath := filepath.Join(tempDir, "key.pk")
	//err = os.WriteFile(tempFilePath, []byte(privateKeyHex), constants.WriteReadUserOnlyPerms)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to write private key to temp file: %w", err)
	//}
	mnemonicId, err := s.CreateKeyMnemonic(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create mnemonic: %v", err)
	}
	// Loop through all key types and import them
	for _, keyType := range keyTypes {
		//output, err := s.ImportKeyRaw(ctx, keyType, tempFilePath)
		//if err != nil {
		//	errors = append(errors, fmt.Errorf("failed to import %s key: %w", keyType, err))
		//	continue
		//}
		//importedKeys = append(importedKeys, output)

		output, err := s.DeriveKeyFromMnemonic(ctx, keyType, mnemonicId)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to import %s key: %w", keyType, err))
			continue
		}
		derivedKeys = append(derivedKeys, output)
	}

	// Temporary files are cleaned up automatically via defer os.RemoveAll(tempDir)

	// Check if any keys were successfully imported
	if len(derivedKeys) == 0 {
		if len(errors) > 0 {
			return nil, status.Errorf(codes.Internal, "failed to import any keys: %v", errors)
		}
		return nil, status.Errorf(codes.Internal, "no keys were imported")
	}

	// TODO: Create account object and return proper response
	// For now, returning a placeholder response
	return &proto.CreateAccountResponse{
		AccountId:     "generated-account-id",
		Addresses:     []string{"placeholder-address"},
		PChainAddress: "placeholder-p-chain-address",
	}, nil
}

// GetAccount retrieves an account by address
func (s *WalletServer) GetAccount(ctx context.Context, req *proto.GetAccountRequest) (*proto.GetAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}

// ListAccounts returns all accounts
func (s *WalletServer) ListAccounts(ctx context.Context, req *emptypb.Empty) (*proto.ListAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}

// ImportAccount imports an existing account
func (s *WalletServer) ImportAccount(ctx context.Context, req *proto.ImportAccountRequest) (*proto.ImportAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportAccount not implemented")
}

// BuildTransaction builds a transaction
func (s *WalletServer) BuildTransaction(ctx context.Context, req *proto.BuildTransactionRequest) (*proto.BuildTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildTransaction not implemented")
}

// SignTransaction signs a transaction
func (s *WalletServer) SignTransaction(ctx context.Context, req *proto.SignTransactionRequest) (*proto.SignTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignTransaction not implemented")
}

// SendTransaction sends a transaction
func (s *WalletServer) SendTransaction(ctx context.Context, req *proto.SendTransactionRequest) (*proto.SendTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTransaction not implemented")
}

// GetChainClients returns chain client endpoints
func (s *WalletServer) GetChainClients(ctx context.Context, req *emptypb.Empty) (*proto.GetChainClientsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChainClients not implemented")
}

// SetChainClients updates chain client endpoints
func (s *WalletServer) SetChainClients(ctx context.Context, req *proto.SetChainClientsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetChainClients not implemented")
}

// Close performs cleanup
func (s *WalletServer) Close(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
