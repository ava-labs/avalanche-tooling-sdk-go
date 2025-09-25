// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/api/generated/api/proto"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

// APIAccount represents an account that communicates with a gRPC server
type APIAccount struct {
	AccountID              string
	ServerAccountAddresses []ids.ShortID
	GrpcClient             proto.WalletServiceClient
}

// Ensure APIAccount implements Account interface
var _ Account = (*APIAccount)(nil)

// NewAPIAccount creates a new API account
func NewAPIAccount(accountID string, addresses []ids.ShortID, grpcClient proto.WalletServiceClient) *APIAccount {
	return &APIAccount{
		AccountID:              accountID,
		ServerAccountAddresses: addresses,
		GrpcClient:             grpcClient,
	}
}

// Addresses returns all addresses associated with this account
func (a *APIAccount) Addresses() []ids.ShortID {
	return a.ServerAccountAddresses
}

// GetPChainAddress returns the P-Chain address for the given network
func (a *APIAccount) GetPChainAddress(network network.Network) (string, error) {
	return "", nil
}

// GetKeychain returns the keychain associated with this account
func (a *APIAccount) GetKeychain() (*secp256k1fx.Keychain, error) {
	// For API accounts, we don't have direct access to the keychain
	// This is a limitation of the API-based approach - the server manages the keys
	// In a real implementation, you might want to:
	// 1. Return a mock keychain with just the addresses
	// 2. Return an error indicating this operation is not supported
	// 3. Implement a different approach for keychain operations

	// For now, we'll return an error indicating this is not supported
	return nil, fmt.Errorf("keychain access not supported for API accounts - keys are managed by the server")
}

// GetAccountID returns the account ID
func (a *APIAccount) GetAccountID() string {
	return a.AccountID
}

// SetAddresses updates the addresses for this account
func (a *APIAccount) SetAddresses(addresses []ids.ShortID) {
	a.ServerAccountAddresses = addresses
}

// NewAccount creates a new account of the same type
func (a *APIAccount) NewAccount() (Account, error) {
	// For API accounts, we can't create new accounts without a gRPC client
	// This is a limitation of the API-based approach
	return nil, fmt.Errorf("NewAccount not supported for API accounts - use wallet.CreateAccount instead")
}
