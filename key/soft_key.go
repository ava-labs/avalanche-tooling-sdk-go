// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package key

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"

	"github.com/ava-labs/avalanchego/utils/formatting/address"

	sdkConstants "github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/cb58"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/libevm/crypto"
)

var (
	ErrInvalidPrivateKey         = errors.New("invalid private key")
	ErrInvalidPrivateKeyLen      = errors.New("invalid private key length (expect 64 bytes in hex)")
	ErrInvalidPrivateKeyEnding   = errors.New("invalid private key ending")
	ErrInvalidPrivateKeyEncoding = errors.New("invalid private key encoding")
)

// SoftKey represents a software-based cryptographic key stored locally on the machine.
// It encapsulates a secp256k1 private key along with its various representations
// and provides functionality for local key management, including generation,
// storage, and retrieval from the local filesystem. SoftKey is designed for
// development, testing, and scenarios where keys are managed directly on the
// local machine rather than through hardware security modules or external
// key management systems.
//
// WARNING: SoftKey is NOT recommended for production use due to security concerns.
// Private keys stored in plaintext on the local filesystem are vulnerable to
// various attack vectors including malware, unauthorized file access, and
// physical device compromise. For production environments, use hardware security
// modules (HSMs), secure key management services, or other secure key storage
// solutions that provide proper key protection and access controls.
//
// This implementation is suitable ONLY for local development environments and
// applications where key security is managed through filesystem permissions
// and local access controls rather than specialized hardware.
type SoftKey struct {
	// privKey is the actual secp256k1 private key used for cryptographic operations
	privKey *secp256k1.PrivateKey
	// privKeyRaw contains raw bytes of the private key for direct access
	privKeyRaw []byte
	// privKeyEncoded is the CB58-encoded string representation for storage/transmission
	privKeyEncoded string
	// keyChain is the Avalanche keychain containing the public key for transaction signing
	keyChain *secp256k1fx.Keychain
}

const (
	privKeyEncPfx = "PrivateKey-"
	privKeySize   = 64
)

type SOp struct {
	privKey        *secp256k1.PrivateKey
	privKeyEncoded string
}

type SOpOption func(*SOp)

func (sop *SOp) applyOpts(opts []SOpOption) {
	for _, opt := range opts {
		opt(sop)
	}
}

// To create a new key SoftKey with a pre-loaded private key.
func WithPrivateKey(privKey *secp256k1.PrivateKey) SOpOption {
	return func(sop *SOp) {
		sop.privKey = privKey
	}
}

// To create a new key SoftKey with a pre-defined private key.
func WithPrivateKeyEncoded(privKey string) SOpOption {
	return func(sop *SOp) {
		sop.privKeyEncoded = privKey
	}
}

func NewSoft(opts ...SOpOption) (*SoftKey, error) {
	ret := &SOp{}
	ret.applyOpts(opts)

	// set via "WithPrivateKeyEncoded"
	if len(ret.privKeyEncoded) > 0 {
		privKey, err := decodePrivateKey(ret.privKeyEncoded)
		if err != nil {
			return nil, err
		}
		// to not overwrite
		if ret.privKey != nil &&
			!bytes.Equal(ret.privKey.Bytes(), privKey.Bytes()) {
			return nil, ErrInvalidPrivateKey
		}
		ret.privKey = privKey
	}

	// generate a new one
	if ret.privKey == nil {
		var err error
		ret.privKey, err = secp256k1.NewPrivateKey()
		if err != nil {
			return nil, err
		}
	}

	privKey := ret.privKey
	privKeyEncoded, err := encodePrivateKey(ret.privKey)
	if err != nil {
		return nil, err
	}

	// double-check encoding is consistent
	if ret.privKeyEncoded != "" &&
		ret.privKeyEncoded != privKeyEncoded {
		return nil, ErrInvalidPrivateKeyEncoding
	}

	keyChain := secp256k1fx.NewKeychain()
	keyChain.Add(privKey)

	m := &SoftKey{
		privKey:        privKey,
		privKeyRaw:     privKey.Bytes(),
		privKeyEncoded: privKeyEncoded,
		keyChain:       keyChain,
	}

	return m, nil
}

// LoadSoft loads the private key from disk and creates the corresponding SoftKey.
func LoadSoft(keyPath string) (*SoftKey, error) {
	kb, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return LoadSoftFromBytes(kb)
}

func LoadSoftOrCreate(keyPath string) (*SoftKey, error) {
	if utils.FileExists(keyPath) {
		return LoadSoft(keyPath)
	} else {
		k, err := NewSoft()
		if err != nil {
			return nil, err
		}
		if err := k.Save(keyPath); err != nil {
			return nil, err
		}
		return k, nil
	}
}

// LoadSoftFromBytes loads the private key from bytes and creates the corresponding SoftKey.
func LoadSoftFromBytes(kb []byte) (*SoftKey, error) {
	// in case, it's already encoded
	k, err := NewSoft(WithPrivateKeyEncoded(string(kb)))
	if err == nil {
		return k, nil
	}

	r := bufio.NewReader(bytes.NewBuffer(kb))
	buf := make([]byte, privKeySize)
	n, err := readASCII(buf, r)
	if err != nil {
		return nil, err
	}
	if n != len(buf) {
		return nil, ErrInvalidPrivateKeyLen
	}
	if err := checkKeyFileEnd(r); err != nil {
		return nil, err
	}

	skBytes, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}
	privKey, err := secp256k1.ToPrivateKey(skBytes)
	if err != nil {
		return nil, err
	}

	return NewSoft(WithPrivateKey(privKey))
}

// readASCII reads into 'buf', stopping when the buffer is full or
// when a non-printable control character is encountered.
func readASCII(buf []byte, r io.ByteReader) (n int, err error) {
	for ; n < len(buf); n++ {
		buf[n], err = r.ReadByte()
		switch {
		case errors.Is(err, io.EOF) || buf[n] < '!':
			return n, nil
		case err != nil:
			return n, err
		}
	}
	return n, nil
}

const fileEndLimit = 1

// checkKeyFileEnd skips over additional newlines at the end of a key file.
func checkKeyFileEnd(r io.ByteReader) error {
	for idx := 0; ; idx++ {
		b, err := r.ReadByte()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		case b != '\n' && b != '\r':
			return ErrInvalidPrivateKeyEnding
		case idx > fileEndLimit:
			return ErrInvalidPrivateKeyLen
		}
	}
}

func encodePrivateKey(pk *secp256k1.PrivateKey) (string, error) {
	privKeyRaw := pk.Bytes()
	enc, err := cb58.Encode(privKeyRaw)
	if err != nil {
		return "", err
	}
	return privKeyEncPfx + enc, nil
}

func decodePrivateKey(enc string) (*secp256k1.PrivateKey, error) {
	rawPk := strings.Replace(enc, privKeyEncPfx, "", 1)
	skBytes, err := cb58.Decode(rawPk)
	if err != nil {
		return nil, err
	}
	privKey, err := secp256k1.ToPrivateKey(skBytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func (m *SoftKey) P(networkHRP string) (string, error) {
	return address.Format("P", networkHRP, m.privKey.PublicKey().Address().Bytes())
}

func (m *SoftKey) X(networkHRP string) (string, error) {
	return address.Format("X", networkHRP, m.privKey.PublicKey().Address().Bytes())
}

func (m *SoftKey) C() string {
	ecdsaPrv := m.privKey.ToECDSA()
	pub := ecdsaPrv.PublicKey

	addr := crypto.PubkeyToAddress(pub)
	return addr.String()
}

// Returns the KeyChain
func (m *SoftKey) KeyChain() *secp256k1fx.Keychain {
	return m.keyChain
}

// Returns the private key encoded hex
func (m *SoftKey) PrivKeyHex() string {
	return hex.EncodeToString(m.privKeyRaw)
}

// Saves the private key to disk with hex encoding.
func (m *SoftKey) Save(p string) error {
	return os.WriteFile(p, []byte(m.PrivKeyHex()), sdkConstants.WriteReadUserOnlyPerms)
}

func (m *SoftKey) Addresses() []ids.ShortID {
	return []ids.ShortID{m.privKey.PublicKey().Address()}
}

func GetHRP(networkID uint32) string {
	switch networkID {
	case constants.LocalID:
		return constants.LocalHRP
	case constants.FujiID:
		return constants.FujiHRP
	case constants.MainnetID:
		return constants.MainnetHRP
	default:
		return constants.FallbackHRP
	}
}

func (m *SoftKey) GetNetworkChainAddress(network network.Network, chain string) ([]string, error) {
	if chain != "P" && chain != "X" {
		return nil, fmt.Errorf("only P or X is accepted as a chain option")
	}
	// Parse HRP to create valid address
	hrp := GetHRP(network.ID)
	addressStr, error := address.Format(chain, hrp, m.privKey.PublicKey().Address().Bytes())
	return []string{addressStr}, error
}
