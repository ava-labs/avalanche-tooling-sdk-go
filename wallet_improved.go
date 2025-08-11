// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/key"
	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain"
	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	utilsSDK "github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/network"
	avagokeychain "github.com/ava-labs/avalanchego/utils/crypto/keychain"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	avmTxs "github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	commonAvago "github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Improved Wallet struct with hidden keychain
type Wallet struct {
	*primary.Wallet
	keychain keychain.Keychain // completely private
	options  []common.Option
	config   primary.WalletConfig
}

// Custom type for private key strings
type PrivateKeyString struct {
	Key string
}

// Single constructor that accepts any key source
func NewWallet(ctx context.Context, network network.Network, keySource interface{}) (*Wallet, error) {
	var keychain keychain.Keychain
	var err error

	switch source := keySource.(type) {
	case string: // File path
		kc, err := keychain.NewKeychain(network, source, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create keychain from file: %w", err)
		}
		keychain = *kc
	case *keychain.LedgerParams: // Ledger
		kc, err := keychain.NewKeychain(network, "", source)
		if err != nil {
			return nil, fmt.Errorf("failed to create keychain from ledger: %w", err)
		}
		keychain = *kc
	case *PrivateKeyString: // Custom type for private key strings
		keychain, err = createKeychainFromPrivateKeyString(network, source.Key)
	default:
		return nil, fmt.Errorf("unsupported key source type: %T", keySource)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create keychain: %w", err)
	}

	// Use default config
	config := primary.WalletConfig{}
	return New(ctx, network.Endpoint, keychain.Keychain, config)
}

// Helper function to create keychain from private key string
func createKeychainFromPrivateKeyString(network network.Network, privateKeyStr string) (keychain.Keychain, error) {
	// Create a temporary file with the private key
	tmpFile, err := os.CreateTemp("", "avalanche-wallet-*.pk")
	if err != nil {
		return keychain.Keychain{}, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up temp file

	// Write private key to temp file
	if _, err := tmpFile.WriteString(privateKeyStr); err != nil {
		return keychain.Keychain{}, fmt.Errorf("failed to write private key to temp file: %w", err)
	}
	tmpFile.Close()

	// Create keychain from temp file
	return keychain.NewKeychain(network, tmpFile.Name(), nil)
}

// Usage examples:
//
// 1. From file path:
//    wallet, err := wallet.NewWallet(ctx, network.FujiNetwork(), "/path/to/key.pk")
//
// 2. From private key string:
//    wallet, err := wallet.NewWallet(ctx, network.FujiNetwork(), &wallet.PrivateKeyString{Key: "PrivateKey-..."})
//
// 3. From ledger:
//    wallet, err := wallet.NewWallet(ctx, network.FujiNetwork(), &keychain.LedgerParams{...})

// Keep the original constructor for advanced use cases
func New(ctx context.Context, uri string, avaxKeychain avagokeychain.Keychain, config primary.WalletConfig) (*Wallet, error) {
	wallet, err := primary.MakeWallet(
		ctx,
		uri,
		avaxKeychain,
		secp256k1fx.NewKeychain(),
		config,
	)
	return &Wallet{
		Wallet: wallet,
		keychain: keychain.Keychain{
			Keychain: avaxKeychain,
		},
		config: config,
	}, err
}

// All keychain operations go through wallet methods (hidden implementation)
func (w *Wallet) Addresses() []ids.ShortID {
	return w.keychain.Addresses().List()
}

func (w *Wallet) PChainAddress(networkHRP string) (string, error) {
	if softKey, ok := w.keychain.Keychain.(*key.SoftKey); ok {
		return softKey.P(networkHRP)
	}
	return "", fmt.Errorf("unsupported keychain type")
}

func (w *Wallet) XChainAddress(networkHRP string) (string, error) {
	if softKey, ok := w.keychain.Keychain.(*key.SoftKey); ok {
		return softKey.X(networkHRP)
	}
	return "", fmt.Errorf("unsupported keychain type")
}

func (w *Wallet) CChainAddress() string {
	if softKey, ok := w.keychain.Keychain.(*key.SoftKey); ok {
		return softKey.C()
	}
	return ""
}

// Convenient transaction signing methods
func (w *Wallet) SignPChainTx(tx *txs.Tx) error {
	return w.P().Signer().Sign(context.Background(), tx)
}

func (w *Wallet) SignXChainTx(tx *avmTxs.Tx) error {
	return w.X().Signer().Sign(context.Background(), tx)
}

func (w *Wallet) SignCChainTx(tx *types.Transaction) error {
	return w.C().Signer().Sign(context.Background(), tx)
}

// Transaction issuing methods
func (w *Wallet) IssuePChainTx(tx *txs.Tx, options ...common.Option) error {
	// Implementation would use the embedded primary.Wallet
	return fmt.Errorf("not implemented yet")
}

func (w *Wallet) IssueXChainTx(tx *avmTxs.Tx, options ...common.Option) error {
	// Implementation would use the embedded primary.Wallet
	return fmt.Errorf("not implemented yet")
}

func (w *Wallet) IssueCChainTx(tx *types.Transaction, options ...common.Option) error {
	// Implementation would integrate with the evm package
	return fmt.Errorf("not implemented yet")
}

// Existing wallet methods (unchanged)
func (w *Wallet) SecureWalletIsChangeOwner() {
	addrs := w.Addresses()
	changeAddr := addrs[0]
	changeOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{changeAddr},
	}
	w.options = append(w.options, common.WithChangeOwner(changeOwner))
	w.Wallet = primary.NewWalletWithOptions(w.Wallet, w.options...)
}

func (w *Wallet) SetAuthKeys(authKeys []ids.ShortID) {
	addrs := w.Addresses()
	addrsSet := set.Set[ids.ShortID]{}
	addrsSet.Add(addrs...)
	addrsSet.Add(authKeys...)
	w.options = append(w.options, common.WithCustomAddresses(addrsSet))
	w.Wallet = primary.NewWalletWithOptions(w.Wallet, w.options...)
}

func (w *Wallet) SetSubnetAuthMultisig(authKeys []ids.ShortID) {
	w.SecureWalletIsChangeOwner()
	w.SetAuthKeys(authKeys)
}

func (w *Wallet) GetMultisigTxOptions(subnetAuthKeys []ids.ShortID) []common.Option {
	options := []common.Option{}
	walletAddrs := w.Addresses()
	changeAddr := walletAddrs[0]
	customAddrsSet := set.Set[ids.ShortID]{}
	customAddrsSet.Add(walletAddrs...)
	customAddrsSet.Add(subnetAuthKeys...)
	options = append(options, common.WithCustomAddresses(customAddrsSet))
	changeOwner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{changeAddr},
	}
	options = append(options, common.WithChangeOwner(changeOwner))
	return options
}

func (w *Wallet) Commit(ms multisig.Multisig, waitForTxAcceptance bool) (ids.ID, error) {
	if ms.Undefined() {
		return ids.Empty, multisig.ErrUndefinedTx
	}
	isReady, err := ms.IsReadyToCommit()
	if err != nil {
		return ids.Empty, err
	}
	if !isReady {
		return ids.Empty, errors.New("tx is not fully signed so can't be committed")
	}
	tx, err := ms.GetWrappedPChainTx()
	if err != nil {
		return ids.Empty, err
	}

	const (
		repeats             = 3
		sleepBetweenRepeats = 2 * time.Second
	)

	var issueTxErr error
	for i := 0; i < repeats; i++ {
		ctx, cancel := utilsSDK.GetAPILargeContext()
		defer cancel()
		options := []commonAvago.Option{commonAvago.WithContext(ctx)}
		if !waitForTxAcceptance {
			options = append(options, commonAvago.WithAssumeDecided())
		}
		issueTxErr = w.IssuePChainTx(tx, options...)
		if issueTxErr == nil {
			break
		}
		if ctx.Err() != nil {
			issueTxErr = fmt.Errorf("timeout issuing/verifying tx with ID %s: %w", tx.ID(), issueTxErr)
		} else {
			issueTxErr = fmt.Errorf("error issuing tx with ID %s: %w", tx.ID(), issueTxErr)
		}
		time.Sleep(sleepBetweenRepeats)
	}
	if issueTxErr != nil {
		return ids.Empty, fmt.Errorf("issue tx error %w", issueTxErr)
	}
	return tx.ID(), issueTxErr
}
