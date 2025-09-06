// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package keychain

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/chain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/pchain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/xchain"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto/keychain"
	avasecp256k1 "github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/wallet/chain/c"

	"github.com/cubist-labs/cubesigner-go-sdk/client"
	"github.com/cubist-labs/cubesigner-go-sdk/models"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ keychain.Keychain = (*cubesignerKeychain)(nil)
	_ c.EthKeychain     = (*cubesignerKeychain)(nil)
	_ keychain.Signer   = (*cubesignerSigner)(nil)
)

type addrInfo struct {
	chain      string
	materialID string
	pubKey     *avasecp256k1.PublicKey
}

// cubesignerKeychain is an abstraction of the underlying cubesigner connection,
// to be able to get a signer from a specific address
type cubesignerKeychain struct {
	cubesignerClient client.ApiClient
	avaAddrs         set.Set[ids.ShortID]
	ethAddrs         set.Set[common.Address]
	avaAddrToInfo    map[ids.ShortID]addrInfo
	ethAddrToInfo    map[common.Address]addrInfo
}

// parseAddress parses an address string and returns keyType, chain, materialID
func parseAddress(addr string) (models.KeyType, string, string, error) {
	if common.IsHexAddress(addr) {
		// ETH address
		return models.SecpEthAddr, "C", addr, nil
	}

	// Avalanche Bech32 address
	addrParts := strings.SplitN(addr, "-", 2)
	if len(addrParts) != 2 {
		return models.KeyType(""), "", "", fmt.Errorf("invalid format for server keychain address \"%s\"", addr)
	}

	_, hrp, _, err := address.Parse(addr)
	if err != nil {
		return models.KeyType(""), "", "", fmt.Errorf("invalid format for server keychain address %s: %w", addr, err)
	}

	var keyType models.KeyType
	switch {
	case hrp == constants.FujiHRP:
		keyType = models.SecpAvaTestAddr
	case hrp == constants.MainnetHRP:
		keyType = models.SecpAvaAddr
	default:
		return models.KeyType(""), "", "", fmt.Errorf("server can't sign addresses with hrp %s", hrp)
	}

	return keyType, addrParts[0], addrParts[1], nil
}

// validateAndProcessKey fetches and validates a public key from cubesigner
func validateAndProcessKey(cubesignerClient client.ApiClient, keyType models.KeyType, materialID, addr string) (*avasecp256k1.PublicKey, error) {
	// Validate key exists
	keyInfo, err := cubesignerClient.GetKeyByMaterialId(string(keyType), materialID)
	if err != nil {
		return nil, fmt.Errorf("could not find server address %s: %w", addr, err)
	}

	// Process public key
	pubKeyHex := strings.TrimPrefix(keyInfo.PublicKey, "0x")
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key for server address %s: %w", addr, err)
	}
	if len(pubKeyBytes) != 65 {
		return nil, fmt.Errorf("invalid public key length for server address %s: expected 65 bytes, got %d", addr, len(pubKeyBytes))
	}
	if pubKeyBytes[0] != 0x04 {
		return nil, fmt.Errorf("invalid public key format for server address %s: expected uncompressed format (0x04 prefix), got 0x%02x", addr, pubKeyBytes[0])
	}

	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid public key format for server address %s: %w", addr, err)
	}

	avaPubKey, err := avasecp256k1.ToPublicKey(pubKey.SerializeCompressed())
	if err != nil {
		return nil, fmt.Errorf("invalid public key format for server address %s: %w", addr, err)
	}

	return avaPubKey, nil
}

// NewCubeSignerKeychain creates a new keychain abstraction over a cubesigner connection
func NewCubesignerKeychain(
	cubesignerClient client.ApiClient,
	addrs []string,
) (keychain.Keychain, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("can't create a server keychain from 0 addresses")
	}

	avaAddrs := set.Set[ids.ShortID]{}
	ethAddrs := set.Set[common.Address]{}
	avaAddrToInfo := map[ids.ShortID]addrInfo{}
	ethAddrToInfo := map[common.Address]addrInfo{}

	for _, addr := range addrs {
		// Parse address to get type and components
		keyType, chain, materialID, err := parseAddress(addr)
		if err != nil {
			return nil, err
		}

		// Validate and process the public key
		avaPubKey, err := validateAndProcessKey(cubesignerClient, keyType, materialID, addr)
		if err != nil {
			return nil, err
		}

		info := addrInfo{
			chain:      chain,
			materialID: materialID,
			pubKey:     avaPubKey,
		}

		// Add to appropriate collections based on address type
		if keyType == models.SecpEthAddr {
			kFromAddr := common.HexToAddress(addr)
			kFromPubKey := avaPubKey.EthAddress()
			if kFromAddr != kFromPubKey {
				return nil, fmt.Errorf("address %s obtained from public key is inconsistent with server address %s", kFromPubKey, kFromAddr)
			}
			ethAddrs.Add(kFromPubKey)
			ethAddrToInfo[kFromPubKey] = info
		} else {
			kFromAddr, err := address.ParseToID(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid format for server keychain address %s: %w", addr, err)
			}
			kFromPubKey := avaPubKey.Address()
			if kFromAddr != kFromPubKey {
				return nil, fmt.Errorf("address %s obtained from public key is inconsistent with server address %s", kFromPubKey, kFromAddr)
			}
			avaAddrs.Add(kFromPubKey)
			avaAddrToInfo[kFromPubKey] = info
		}
	}

	return &cubesignerKeychain{
		cubesignerClient: cubesignerClient,
		avaAddrs:         avaAddrs,
		ethAddrs:         ethAddrs,
		avaAddrToInfo:    avaAddrToInfo,
		ethAddrToInfo:    ethAddrToInfo,
	}, nil
}

func (kc *cubesignerKeychain) Addresses() set.Set[ids.ShortID] {
	return kc.avaAddrs
}

func (kc *cubesignerKeychain) Get(addr ids.ShortID) (keychain.Signer, bool) {
	info, found := kc.avaAddrToInfo[addr]
	if !found {
		return nil, false
	}
	return &cubesignerSigner{
		cubesignerClient: kc.cubesignerClient,
		info:             info,
	}, true
}

func (kc *cubesignerKeychain) EthAddresses() set.Set[common.Address] {
	return kc.ethAddrs
}

func (kc *cubesignerKeychain) GetEth(addr common.Address) (keychain.Signer, bool) {
	info, found := kc.ethAddrToInfo[addr]
	if !found {
		return nil, false
	}
	return &cubesignerSigner{
		cubesignerClient: kc.cubesignerClient,
		info:             info,
	}, true
}

// cubesignerAvagoSigner is an abstraction of the underlying cubesigner connection,
// to be able sign for a specific address
type cubesignerSigner struct {
	cubesignerClient client.ApiClient
	info             addrInfo
}

// expects to receive a hash of the unsigned tx bytes
func (*cubesignerSigner) SignHash([]byte) ([]byte, error) {
	return nil, fmt.Errorf("server does not support signing a tx hash")
}

// expects to receive the unsigned tx bytes
func (s *cubesignerSigner) Sign(b []byte) ([]byte, error) {
	// Get chain configuration based on transaction bytes
	chainStr, materialID, err := s.getChainAndMaterialID(b)
	if err != nil {
		return nil, err
	}

	response, err := s.cubesignerClient.AvaSerializedTxSign(
		chainStr,
		materialID,
		models.AvaSerializedTxSignRequest{
			Tx: "0x" + hex.EncodeToString(b),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("server signing err: %w", err)
	}
	if response.ResponseData == nil {
		return nil, fmt.Errorf("empty signature obtained from server")
	}
	signatureBytes, err := hex.DecodeString(response.ResponseData.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode server's signature: %w", err)
	}
	if len(signatureBytes) != avasecp256k1.SignatureLen {
		return nil, fmt.Errorf("invalid server's signature length: expected %d bytes, got %d", avasecp256k1.SignatureLen, len(signatureBytes))
	}
	return signatureBytes, nil
}

// getMaterialIDForBech32Chain extracts HRP from transaction and formats address as Bech32
func (s *cubesignerSigner) getMaterialIDForBech32Chain(b []byte, chainName string) (string, error) {
	var hrp string
	var err error

	switch chainName {
	case "X":
		xTx, _ := xchain.TxFromBytes(b) // We know this succeeds from the guard
		hrp, err = xchain.GetHRP(xTx)
		if err != nil {
			return "", fmt.Errorf("failed to extract HRP from X-Chain transaction: %w", err)
		}
	case "P":
		pTx, _ := pchain.TxFromBytes(b) // We know this succeeds from the guard
		hrp, err = pchain.GetHRP(pTx)
		if err != nil {
			return "", fmt.Errorf("failed to extract HRP from P-Chain transaction: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported chain for Bech32 formatting: %s", chainName)
	}

	addr := s.info.pubKey.Address()
	materialID, err := address.FormatBech32(hrp, addr.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format %s address as Bech32: %w", chainName, err)
	}

	return materialID, nil
}

// getChainAndMaterialID determines the appropriate chain string and material ID based on transaction bytes
func (s *cubesignerSigner) getChainAndMaterialID(b []byte) (string, string, error) {
	// Start with defaults
	chainStr := s.info.chain
	materialID := s.info.materialID

	// Detect chain type using safeguard logic
	chainType := chain.DetectTxChainType(b)

	// If chain type is clearly identified, use chain-specific logic
	switch chainType {
	case chain.CChain:
		chainStr = "C"
		materialID = s.info.pubKey.EthAddress().Hex()
	case chain.XChain:
		chainStr = "X"
		var err error
		materialID, err = s.getMaterialIDForBech32Chain(b, "X")
		if err != nil {
			return chainStr, materialID, err
		}
	case chain.PChain:
		chainStr = "P"
		var err error
		materialID, err = s.getMaterialIDForBech32Chain(b, "P")
		if err != nil {
			return chainStr, materialID, err
		}
	case chain.Undefined:
		// Keep default values for undefined/ambiguous cases
	}

	return chainStr, materialID, nil
}

func (s *cubesignerSigner) Address() ids.ShortID {
	return s.info.pubKey.Address()
}
