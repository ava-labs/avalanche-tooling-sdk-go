// Copyright (C) 222, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package multisig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

type TxKind int64

const (
	Undefined TxKind = iota
	PChainRemoveSubnetValidatorTx
	PChainAddSubnetValidatorTx
	PChainCreateChainTx
	PChainTransformSubnetTx
	PChainAddPermissionlessValidatorTx
	PChainTransferSubnetOwnershipTx
)

type Multisig struct {
	pChainTx    *txs.Tx
	controlKeys []ids.ShortID
	threshold   uint32
}

var ErrNoRemainingAuthSignersInWallet = errors.New("wallet does not contain any renaububg auth signer")

func New(pChainTx *txs.Tx) *Multisig {
	ms := Multisig{
		pChainTx: pChainTx,
	}
	return &ms
}

func (ms *Multisig) String() string {
	if ms.pChainTx != nil {
		return ms.pChainTx.ID().String()
	}
	return ""
}

func (ms *Multisig) ToBytes() ([]byte, error) {
	if ms.pChainTx != nil {
		txBytes, err := txs.Codec.Marshal(txs.CodecVersion, ms.pChainTx)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal signed tx: %w", err)
		}
		return txBytes, nil
	}
	return nil, fmt.Errorf("undefined tx")
}

func (ms *Multisig) ToFile(txPath string) error {
	if ms.pChainTx != nil {
		txBytes, err := ms.ToBytes()
		if err != nil {
			return err
		}
		txStr, err := formatting.Encode(formatting.Hex, txBytes)
		if err != nil {
			return fmt.Errorf("couldn't encode signed tx: %w", err)
		}
		f, err := os.Create(txPath)
		if err != nil {
			return fmt.Errorf("couldn't create file to write tx to: %w", err)
		}
		defer f.Close()
		_, err = f.WriteString(txStr)
		if err != nil {
			return fmt.Errorf("couldn't write tx into file: %w", err)
		}
		return nil
	}
	return fmt.Errorf("undefined tx")
}

func (ms *Multisig) FromBytes(txBytes []byte) error {
	var tx txs.Tx
	if _, err := txs.Codec.Unmarshal(txBytes, &tx); err != nil {
		return fmt.Errorf("error unmarshaling signed tx: %w", err)
	}
	if err := tx.Initialize(txs.Codec); err != nil {
		return fmt.Errorf("error initializing signed tx: %w", err)
	}
	ms.pChainTx = &tx
	return nil
}

func (ms *Multisig) FromFile(txPath string) error {
	txEncodedBytes, err := os.ReadFile(txPath)
	if err != nil {
		return err
	}
	txBytes, err := formatting.Decode(formatting.Hex, string(txEncodedBytes))
	if err != nil {
		return fmt.Errorf("couldn't decode signed tx: %w", err)
	}
	return ms.FromBytes(txBytes)
}

func (ms *Multisig) Commit(wallet wallet.Wallet, waitForTxAcceptance bool) (ids.ID, error) {
	const (
		repeats             = 3
		sleepBetweenRepeats = 2 * time.Second
	)
	var issueTxErr error
	for i := 0; i < repeats; i++ {
		ctx, cancel := utils.GetAPILargeContext()
		defer cancel()
		options := []common.Option{common.WithContext(ctx)}
		if !waitForTxAcceptance {
			options = append(options, common.WithAssumeDecided())
		}
		// TODO: split error checking and recovery between issuing and waiting for status
		issueTxErr = wallet.P().IssueTx(ms.pChainTx, options...)
		if issueTxErr == nil {
			break
		}
		if ctx.Err() != nil {
			issueTxErr = fmt.Errorf("timeout issuing/verifying tx with ID %s: %w", ms.pChainTx.ID(), issueTxErr)
		} else {
			issueTxErr = fmt.Errorf("error issuing tx with ID %s: %w", ms.pChainTx.ID(), issueTxErr)
		}
		time.Sleep(sleepBetweenRepeats)
	}
	// TODO: having a commit error, maybe should be useful to reestart the wallet internal info
	return ms.pChainTx.ID(), issueTxErr
}

func (ms *Multisig) IsReadyToCommit() (bool, error) {
	if ms.pChainTx != nil {
		_, remainingSigners, err := ms.GetRemainingAuthSigners()
		if err != nil {
			return false, err
		}
		return len(remainingSigners) == 0, nil
	}
	return false, nil
}

// get subnet auth addresses that did not yet signed a given tx
//   - get the string slice of auth signers for the tx (GetAuthSigners)
//   - verifies that all creds in tx.Creds, except the last one, are fully signed
//     (a cred is fully signed if all the signatures in cred.Sigs are non-empty)
//   - computes remaining signers by iterating the last cred in tx.Creds, associated to subnet auth signing
//   - for each sig in cred.Sig: if sig is empty, then add the associated auth signer address (obtained from
//     authSigners by using the index) to the remaining signers list
//
// if the tx is fully signed, returns empty slice
func (ms *Multisig) GetRemainingAuthSigners() ([]ids.ShortID, []ids.ShortID, error) {
	if ms.pChainTx != nil {
		authSigners, err := ms.GetAuthSigners()
		if err != nil {
			return nil, nil, err
		}
		emptySig := [secp256k1.SignatureLen]byte{}
		// we should have at least 1 cred for output owners and 1 cred for subnet auth
		if len(ms.pChainTx.Creds) < 2 {
			return nil, nil, fmt.Errorf("expected tx.Creds of len 2, got %d", len(ms.pChainTx.Creds))
		}
		// signatures for output owners should be filled (all creds except last one)
		for credIndex := range ms.pChainTx.Creds[:len(ms.pChainTx.Creds)-1] {
			cred, ok := ms.pChainTx.Creds[credIndex].(*secp256k1fx.Credential)
			if !ok {
				return nil, nil, fmt.Errorf("expected cred to be of type *secp256k1fx.Credential, got %T", ms.pChainTx.Creds[credIndex])
			}
			for i, sig := range cred.Sigs {
				if sig == emptySig {
					return nil, nil, fmt.Errorf("expected funding sig %d of cred %d to be filled", i, credIndex)
				}
			}
		}
		// signatures for subnet auth (last cred)
		cred, ok := ms.pChainTx.Creds[len(ms.pChainTx.Creds)-1].(*secp256k1fx.Credential)
		if !ok {
			return nil, nil, fmt.Errorf("expected cred to be of type *secp256k1fx.Credential, got %T", ms.pChainTx.Creds[1])
		}
		if len(cred.Sigs) != len(authSigners) {
			return nil, nil, fmt.Errorf("expected number of cred's signatures %d to equal number of auth signers %d",
				len(cred.Sigs),
				len(authSigners),
			)
		}
		remainingSigners := []ids.ShortID{}
		for i, sig := range cred.Sigs {
			if sig == emptySig {
				remainingSigners = append(remainingSigners, authSigners[i])
			}
		}
		return authSigners, remainingSigners, nil
	}
	return nil, nil, fmt.Errorf("undefined tx")
}

// get all subnet auth addresses that are required to sign a given tx
//   - get subnet control keys as string slice using P-Chain API (GetOwners)
//   - get subnet auth indices from the tx, field tx.UnsignedTx.SubnetAuth
//   - creates the string slice of required subnet auth addresses by applying
//     the indices to the control keys slice
func (ms *Multisig) GetAuthSigners() ([]ids.ShortID, error) {
	if ms.pChainTx != nil {
		controlKeys, _, err := ms.GetSubnetOwners()
		if err != nil {
			return nil, err
		}
		unsignedTx := ms.pChainTx.Unsigned
		var subnetAuth verify.Verifiable
		switch unsignedTx := unsignedTx.(type) {
		case *txs.RemoveSubnetValidatorTx:
			subnetAuth = unsignedTx.SubnetAuth
		case *txs.AddSubnetValidatorTx:
			subnetAuth = unsignedTx.SubnetAuth
		case *txs.CreateChainTx:
			subnetAuth = unsignedTx.SubnetAuth
		case *txs.TransformSubnetTx:
			subnetAuth = unsignedTx.SubnetAuth
		case *txs.TransferSubnetOwnershipTx:
			subnetAuth = unsignedTx.SubnetAuth
		default:
			return nil, fmt.Errorf("unexpected unsigned tx type %T", unsignedTx)
		}
		subnetInput, ok := subnetAuth.(*secp256k1fx.Input)
		if !ok {
			return nil, fmt.Errorf("expected subnetAuth of type *secp256k1fx.Input, got %T", subnetAuth)
		}
		authSigners := []ids.ShortID{}
		for _, sigIndex := range subnetInput.SigIndices {
			if sigIndex >= uint32(len(controlKeys)) {
				return nil, fmt.Errorf("signer index %d exceeds number of control keys", sigIndex)
			}
			authSigners = append(authSigners, controlKeys[sigIndex])
		}
		return authSigners, nil
	}
	return nil, fmt.Errorf("undefined tx")
}

func (*Multisig) GetSpendSigners() ([]ids.ShortID, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (ms *Multisig) GetTxKind() (TxKind, error) {
	if ms.pChainTx != nil {
		unsignedTx := ms.pChainTx.Unsigned
		switch unsignedTx := unsignedTx.(type) {
		case *txs.RemoveSubnetValidatorTx:
			return PChainRemoveSubnetValidatorTx, nil
		case *txs.AddSubnetValidatorTx:
			return PChainAddSubnetValidatorTx, nil
		case *txs.CreateChainTx:
			return PChainCreateChainTx, nil
		case *txs.TransformSubnetTx:
			return PChainTransformSubnetTx, nil
		case *txs.AddPermissionlessValidatorTx:
			return PChainAddPermissionlessValidatorTx, nil
		case *txs.TransferSubnetOwnershipTx:
			return PChainTransferSubnetOwnershipTx, nil
		default:
			return Undefined, fmt.Errorf("unexpected unsigned tx type %T", unsignedTx)
		}
	}
	return Undefined, fmt.Errorf("undefined tx")
}

// get network id associated to tx
func (ms *Multisig) GetNetworkID() (uint32, error) {
	if ms.pChainTx != nil {
		unsignedTx := ms.pChainTx.Unsigned
		var networkID uint32
		switch unsignedTx := unsignedTx.(type) {
		case *txs.RemoveSubnetValidatorTx:
			networkID = unsignedTx.NetworkID
		case *txs.AddSubnetValidatorTx:
			networkID = unsignedTx.NetworkID
		case *txs.CreateChainTx:
			networkID = unsignedTx.NetworkID
		case *txs.TransformSubnetTx:
			networkID = unsignedTx.NetworkID
		case *txs.AddPermissionlessValidatorTx:
			networkID = unsignedTx.NetworkID
		case *txs.TransferSubnetOwnershipTx:
			networkID = unsignedTx.NetworkID
		default:
			return 0, fmt.Errorf("unexpected unsigned tx type %T", unsignedTx)
		}
		return networkID, nil
	}
	return 0, fmt.Errorf("undefined tx")
}

// get network model associated to tx
func (ms *Multisig) GetNetwork() (avalanche.Network, error) {
	if ms.pChainTx != nil {
		networkID, err := ms.GetNetworkID()
		if err != nil {
			return avalanche.UndefinedNetwork, err
		}
		network := avalanche.NetworkFromNetworkID(networkID)
		if network.Kind == avalanche.Undefined {
			return avalanche.UndefinedNetwork, fmt.Errorf("undefined network model for tx")
		}
		return network, nil
	}
	return avalanche.Network{}, fmt.Errorf("undefined tx")
}

func (ms *Multisig) GetBlockchainID() (ids.ID, error) {
	if ms.pChainTx != nil {
		unsignedTx := ms.pChainTx.Unsigned
		var blockchainID ids.ID
		switch unsignedTx := unsignedTx.(type) {
		case *txs.RemoveSubnetValidatorTx:
			blockchainID = unsignedTx.BlockchainID
		case *txs.AddSubnetValidatorTx:
			blockchainID = unsignedTx.BlockchainID
		case *txs.CreateChainTx:
			blockchainID = unsignedTx.BlockchainID
		case *txs.TransformSubnetTx:
			blockchainID = unsignedTx.BlockchainID
		case *txs.AddPermissionlessValidatorTx:
			blockchainID = unsignedTx.BlockchainID
		case *txs.TransferSubnetOwnershipTx:
			blockchainID = unsignedTx.BlockchainID
		default:
			return ids.Empty, fmt.Errorf("unexpected unsigned tx type %T", unsignedTx)
		}
		return blockchainID, nil
	}
	return ids.Empty, fmt.Errorf("undefined tx")
}

// get subnet id associated to tx
func (ms *Multisig) GetSubnetID() (ids.ID, error) {
	if ms.pChainTx != nil {
		unsignedTx := ms.pChainTx.Unsigned
		var subnetID ids.ID
		switch unsignedTx := unsignedTx.(type) {
		case *txs.RemoveSubnetValidatorTx:
			subnetID = unsignedTx.Subnet
		case *txs.AddSubnetValidatorTx:
			subnetID = unsignedTx.SubnetValidator.Subnet
		case *txs.CreateChainTx:
			subnetID = unsignedTx.SubnetID
		case *txs.TransformSubnetTx:
			subnetID = unsignedTx.Subnet
		case *txs.AddPermissionlessValidatorTx:
			subnetID = unsignedTx.Subnet
		case *txs.TransferSubnetOwnershipTx:
			subnetID = unsignedTx.Subnet
		default:
			return ids.Empty, fmt.Errorf("unexpected unsigned tx type %T", unsignedTx)
		}
		return subnetID, nil
	}
	return ids.Empty, fmt.Errorf("undefined tx")
}

func (ms *Multisig) GetSubnetOwners() ([]ids.ShortID, uint32, error) {
	if ms.controlKeys == nil {
		subnetID, err := ms.GetSubnetID()
		if err != nil {
			return nil, 0, err
		}
		network, err := ms.GetNetwork()
		if err != nil {
			return nil, 0, err
		}
		controlKeys, threshold, err := GetOwners(network, subnetID)
		if err != nil {
			return nil, 0, err
		}
		ms.controlKeys = controlKeys
		ms.threshold = threshold
	}
	return ms.controlKeys, ms.threshold, nil
}

func GetOwners(_ avalanche.Network, _ ids.ID) ([]ids.ShortID, uint32, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (ms *Multisig) Sign(
	wallet wallet.Wallet,
	checkAuth bool,
	commitIfReady bool,
	waitForTxAcceptanceOnCommit bool,
) (bool, bool, ids.ID, error) {
	if ms.pChainTx != nil {
		if checkAuth {
			remainingInWallet, err := ms.GetRemainingAuthSignersInWallet(wallet)
			if err != nil {
				return false, false, ids.Empty, fmt.Errorf("error signing tx: %w", err)
			}
			if len(remainingInWallet) == 0 {
				return false, false, ids.Empty, ErrNoRemainingAuthSignersInWallet
			}
		}
		if err := wallet.P().Signer().Sign(context.Background(), ms.pChainTx); err != nil {
			return false, false, ids.Empty, fmt.Errorf("error signing tx: %w", err)
		}
		isReady, err := ms.IsReadyToCommit()
		if err != nil {
			return false, false, ids.Empty, err
		}
		if commitIfReady && isReady {
			_, err := ms.Commit(wallet, waitForTxAcceptanceOnCommit)
			return isReady, err == nil, ms.pChainTx.ID(), err
		}
		return isReady, false, ms.pChainTx.ID(), nil
	}
	return false, false, ids.Empty, fmt.Errorf("undefined tx")
}

func (ms *Multisig) GetRemainingAuthSignersInWallet(wallet wallet.Wallet) ([]ids.ShortID, error) {
	_, subnetAuth, err := ms.GetRemainingAuthSigners()
	if err != nil {
		return nil, err
	}
	walletAddrs := wallet.Addresses()
	subnetAuthInWallet := []ids.ShortID{}
	for _, walletAddr := range walletAddrs {
		for _, addr := range subnetAuth {
			if addr == walletAddr {
				subnetAuthInWallet = append(subnetAuthInWallet, addr)
			}
		}
	}
	return subnetAuthInWallet, nil
}
