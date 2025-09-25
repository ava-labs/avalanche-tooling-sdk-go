// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/models"
	"github.com/cubist-labs/cubesigner-go-sdk/session"

	"github.com/ava-labs/avalanche-tooling-sdk-go/account"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	fujiKeyType          = models.SecpAvaTestAddr
	mainnetKeyType       = models.SecpAvaAddr
	ethKeytype           = models.SecpEthAddr
	avaKeyDerivationPath = "m/44'/9000'/0'/0/0"
	ethKeyDerivationPath = "m/44'/60'/0'/0/0"
	accountsDataFile     = "data/accounts.json"
)

// AccountStorage represents the JSON file structure with address as key
type AccountStorage struct {
	Accounts map[string]models.KeyInfo `json:"accounts"`
}

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
	accounts  []account.Account
	apiClient *client.ApiClient
}

func (s *WalletServer) DeriveKeyFromMnemonic(keyType models.KeyType, mnemonicId string) (models.KeyInfo, error) {
	derivationPath := ""
	if keyType == "" {
		return models.KeyInfo{}, fmt.Errorf("key type cannot be empty")
	}
	if keyType == fujiKeyType || keyType == mainnetKeyType {
		derivationPath = avaKeyDerivationPath
	} else if keyType == ethKeytype {
		derivationPath = ethKeyDerivationPath
	}
	keysToDerive := models.KeyTypeAndDerivationPath{
		KeyType:        keyType,
		DerivationPath: derivationPath,
	}
	deriveKeyRequest := models.DeriveKeysRequest{
		MnemonicId:                 &mnemonicId,
		KeyTypesAndDerivationPaths: []models.KeyTypeAndDerivationPath{keysToDerive},
	}
	deriveKeyResp, err := s.apiClient.DeriveKey(deriveKeyRequest)
	if err != nil {
		return models.KeyInfo{}, err
	}

	fmt.Printf("deriveKeyResp %s \n", deriveKeyResp.Keys)
	if len(deriveKeyResp.Keys) > 0 {
		return deriveKeyResp.Keys[0], nil
	}
	return models.KeyInfo{}, fmt.Errorf("no keys were derived")
}

// NewWalletServer creates a new WalletServer
func NewWalletServer() (*WalletServer, error) {
	filePath := "/Users/raymondsukanto/Desktop/management-session.json"
	manager, err := session.NewJsonSessionManager(&filePath)
	if err != nil {
		return nil, err
	}
	apiClient, err := client.NewApiClient(manager)
	if err != nil {
		return nil, err
	}

	return &WalletServer{
		accounts:  []account.Account{},
		apiClient: apiClient,
	}, nil
}

// loadAccountsFromFile loads accounts from the JSON file
func (s *WalletServer) loadAccountsFromFile() (*AccountStorage, error) {
	// Check if file exists
	if _, err := os.Stat(accountsDataFile); os.IsNotExist(err) {
		// File doesn't exist, return empty storage
		return &AccountStorage{Accounts: make(map[string]models.KeyInfo)}, nil
	}

	// Read file
	data, err := os.ReadFile(accountsDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	// Parse JSON
	var storage AccountStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("failed to parse accounts file: %w", err)
	}

	// Initialize map if nil
	if storage.Accounts == nil {
		storage.Accounts = make(map[string]models.KeyInfo)
	}

	return &storage, nil
}

// saveAccountsToFile saves accounts to the JSON file
func (s *WalletServer) saveAccountsToFile(storage *AccountStorage) error {
	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(accountsDataFile), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(accountsDataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	return nil
}

// addAccountToStorage adds new accounts to the storage and saves to file
func (s *WalletServer) addAccountToStorage(accounts map[string]models.KeyInfo) error {
	// Load existing accounts
	storage, err := s.loadAccountsFromFile()
	if err != nil {
		return fmt.Errorf("failed to load existing accounts: %w", err)
	}

	// Add new accounts to the map
	for address, keyInfo := range accounts {
		storage.Accounts[address] = keyInfo
	}

	// Save back to file
	if err := s.saveAccountsToFile(storage); err != nil {
		return fmt.Errorf("failed to save accounts: %w", err)
	}

	return nil
}

// getAccountByAddress retrieves account info by address from storage
func (s *WalletServer) getAccountByAddress(address string) (*models.KeyInfo, error) {
	storage, err := s.loadAccountsFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load accounts: %w", err)
	}

	if keyInfo, exists := storage.Accounts[address]; exists {
		return &keyInfo, nil
	}

	return nil, fmt.Errorf("account not found for address: %s", address)
}

// CreateAccount creates a new account
func (s *WalletServer) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.CreateAccountResponse, error) {
	// Define all key types to iterate through
	keyTypes := []models.KeyType{
		fujiKeyType,    // "secp-ava-test"
		mainnetKeyType, // "secp-ava"
		ethKeytype,     // "secp"
	}

	var derivedKeys map[models.KeyType]models.KeyInfo
	var errors []error

	createKeyRequest := models.CreateKeyRequest{
		KeyType: models.Mnemonic,
	}
	createKeyResp, err := s.apiClient.CreateKey(createKeyRequest)
	if err != nil {
		return nil, err
	}
	materialID := ""
	if len(createKeyResp.Keys) > 0 {
		materialID = createKeyResp.Keys[0].MaterialId
	}

	// Loop through all key types and import them
	for _, keyType := range keyTypes {
		derivedKey, err := s.DeriveKeyFromMnemonic(keyType, materialID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to derive %s key: %w", keyType, err))
			continue
		}
		derivedKeys[keyType] = derivedKey
	}

	// Temporary files are cleaned up automatically via defer os.RemoveAll(tempDir)

	// Check if any keys were successfully imported
	if len(errors) > 0 {
		return nil, status.Errorf(codes.Internal, "failed to import any keys: %v", errors)
	}

	// Create response
	response := &proto.CreateAccountResponse{
		FujiAvaxAddress: derivedKeys[fujiKeyType].MaterialId,
		AvaxAddress:     derivedKeys[mainnetKeyType].MaterialId,
		EthAddress:      derivedKeys[ethKeytype].MaterialId,
		FujiAvaxKeyId:   derivedKeys[fujiKeyType].KeyId,
		AvaxKeyId:       derivedKeys[mainnetKeyType].KeyId,
		EthKeyId:        derivedKeys[ethKeytype].KeyId,
	}

	// Store account data in JSON file before returning
	// Create a map with address as key and key info as value
	accountsToStore := make(map[string]models.KeyInfo)
	accountsToStore[derivedKeys[fujiKeyType].MaterialId] = derivedKeys[fujiKeyType]
	accountsToStore[derivedKeys[mainnetKeyType].MaterialId] = derivedKeys[mainnetKeyType]
	accountsToStore[derivedKeys[ethKeytype].MaterialId] = derivedKeys[ethKeytype]

	// Save to JSON file
	if err := s.addAccountToStorage(accountsToStore); err != nil {
		return response, fmt.Errorf("Warning: failed to save account to storage: %v\n", err)
	}

	return response, nil
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
