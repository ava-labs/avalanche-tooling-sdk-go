// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"golang.org/x/crypto/hkdf"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/models"
)

// ServerAccount represents a local account implementation
type ServerAccount struct {
	cubistKey models.KeyInfo
}

func encryptWithP384(privKeyBytes []byte, clientPrivKey *ecdsa.PrivateKey, serverPubKeyB64 string) (string, error) {
	// 1. Decode server's P384 public key
	serverPubKeyBytes, err := base64.StdEncoding.DecodeString(serverPubKeyB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode server public key: %w", err)
	}

	// 2. Parse server's public key
	x, y := elliptic.Unmarshal(elliptic.P384(), serverPubKeyBytes)
	if x == nil || y == nil {
		return "", fmt.Errorf("failed to parse server public key")
	}

	serverPubKey := &ecdsa.PublicKey{
		Curve: elliptic.P384(),
		X:     x,
		Y:     y,
	}

	// 3. Perform ECDH to get shared secret
	sharedX, sharedY := elliptic.P384().ScalarMult(serverPubKey.X, serverPubKey.Y, clientPrivKey.D.Bytes())
	sharedSecret := elliptic.Marshal(elliptic.P384(), sharedX, sharedY)

	// 4. Generate salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 5. Derive encryption key using HKDF
	hash := sha256.New
	hkdf := hkdf.New(hash, sharedSecret, salt, nil)

	encryptionKey := make([]byte, 32) // AES-256 key
	if _, err := hkdf.Read(encryptionKey); err != nil {
		return "", fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// 6. Encrypt the private key using AES-GCM
	encryptedKey, err := encryptAESGCM(privKeyBytes, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// 7. Return base64 encoded encrypted key
	return base64.StdEncoding.EncodeToString(encryptedKey), nil
}

func encryptAESGCM(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// NewServerAccount creates a new ServerAccount
func NewServerAccount(cubistClient client.ApiClient) (Account, error) {
	//chainID := int64(43113)
	// Create an Avalanche testnet key
	//createKeyRequest := models.CreateKeyRequest{
	//	KeyType: models.SecpAvaTestAddr,
	//	Count:   1,
	//	ChainId: &chainID, // Avalanche Fuji testnet chain ID
	//}
	//
	//// Execute the key creation
	//keysInfo, err := cubistClient.CreateKey(createKeyRequest)
	//if err != nil {
	//	// handle error
	//}
	//return &ServerAccount{
	//	cubistKey: keysInfo.Keys[0],
	//}, nil

	newKey, err := secp256k1.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("newKey err %s \n", err)
	}
	// Your Avalanche key (secp256k1) - the data to import
	privKeyBytes := newKey.Bytes()

	// Get server's P384 encryption key
	importKeyResp, err := cubistClient.CreateKeyImportKey()
	if err != nil {
		return nil, fmt.Errorf("importKeyResp err %s \n", err)
	}
	// Generate P384 key for encryption (NOT secp256k1)
	clientPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("GenerateKey err %s \n", err)
	}
	// Use P384 to encrypt your secp256k1 key
	encryptedKey, err := encryptWithP384(privKeyBytes, clientPrivKey, importKeyResp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("encryptWithP384 err %s \n", err)
	}
	// Encode client public key
	clientPubKeyBytes := elliptic.Marshal(elliptic.P384(), clientPrivKey.PublicKey.X, clientPrivKey.PublicKey.Y)
	clientPubKeyEncoded := base64.StdEncoding.EncodeToString(clientPubKeyBytes)

	// Generate salt (you'll need to store this and use it in the import request)
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Create import request
	importRequest := &models.ImportKeyRequest{
		DkEnc:     importKeyResp.DkEnc,
		PublicKey: importKeyResp.PublicKey,
		SkEnc:     importKeyResp.SkEnc,
		Expires:   importKeyResp.Expires,
		KeyType:   models.SecpAvaAddr,
		KeyMaterial: []models.ImportKeyRequestMaterial{
			{
				ClientPublicKey: clientPubKeyEncoded,
				IkmEnc:          encryptedKey,
				Salt:            base64.StdEncoding.EncodeToString(salt),
			},
		},
	}

	importKeyResponse, err := cubistClient.ImportKey(*importRequest)
	if err != nil {
		return nil, fmt.Errorf("ImportKey err %s \n", err)
	}
	fmt.Printf("importKeyResponse %s \n", importKeyResponse.Keys)
	return &ServerAccount{}, nil
}

//
//func Import(keyPath string) (Account, error) {
//	k, err := key.LoadSoft(keyPath)
//	if err != nil {
//		return nil, err
//	}
//	return &LocalAccount{
//		SoftKey: k,
//	}, nil
//}

// Addresses returns all addresses associated with this local account
func (a *ServerAccount) Addresses() []ids.ShortID {
	//if a.SoftKey == nil {
	//	return []ids.ShortID{}
	//}
	//return a.SoftKey.KeyChain().Addresses().List()
	return nil
}

func (a *ServerAccount) GetPChainAddress(network network.Network) (string, error) {
	//if a.SoftKey == nil {
	//	return "", fmt.Errorf("SoftKey not initialized")
	//}
	//pchainAddrs, err := a.SoftKey.GetNetworkChainAddress(network, "P")
	//return pchainAddrs[0], err
	return "", nil
}

func (a *ServerAccount) GetKeychain() (*secp256k1fx.Keychain, error) {
	//if a.SoftKey == nil {
	//	return nil, fmt.Errorf("SoftKey not initialized")
	//}
	//return a.SoftKey.KeyChain(), nil
	return nil, fmt.Errorf("GetKeychain not implemented")
}
