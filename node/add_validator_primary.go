// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"golang.org/x/net/context"
)

type PrimaryNetworkValidatorParams struct {
	// NodeID is the unique identifier of the node to be added as a validator on the Primary Network.
	NodeID ids.NodeID

	// Duration is how long the node will be staking the Primary Network
	// Duration has to be greater than or equal to minimum duration for the specified network
	// (Fuji / Mainnet)
	Duration time.Duration

	// StakeAmount is the amount of Avalanche tokens (AVAX) to stake in this validator
	// StakeAmount is in the amount of nAVAX
	// StakeAmount has to be greater than or equal to minimum stake required for the specified network
	StakeAmount uint64

	// DelegationFee is the percent fee this validator will charge when others delegate stake to it
	// When DelegationFee is not set, the minimum delegation fee for the specified network will be set
	// For more information on delegation fee, please head to https://docs.avax.network/nodes/validate/node-validator#delegation-fee-rate
	DelegationFee uint32
}

// ValidatePrimaryNetwork adds node as primary network validator.
// It adds the node in the specified network (Fuji / Mainnet / Devnet)
// and uses the wallet provided in the argument to pay for the transaction fee
func (h *Node) ValidatePrimaryNetwork(
	network avalanche.Network,
	validator PrimaryNetworkValidatorParams,
	wallet wallet.Wallet,
) (ids.ID, error) {
	if validator.NodeID == ids.EmptyNodeID {
		return ids.Empty, subnet.ErrEmptyValidatorNodeID
	}

	if validator.Duration == 0 {
		return ids.Empty, subnet.ErrEmptyValidatorDuration
	}

	minValStake, err := GetMinStakingAmount(network)
	if err != nil {
		return ids.Empty, err
	}

	if validator.StakeAmount < minValStake {
		return ids.Empty, fmt.Errorf("invalid weight, must be greater than or equal to %d: %d", minValStake, validator.StakeAmount)
	}

	if validator.DelegationFee == 0 {
		validator.DelegationFee = network.GenesisParams().MinDelegationFee
	}

	if err = h.HandleBLSKey(); err != nil {
		return ids.Empty, fmt.Errorf("unable to set BLS key of node due to %w", err)
	}

	wallet.SetSubnetAuthMultisig([]ids.ShortID{})

	owner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs: []ids.ShortID{
			wallet.Addresses()[0],
		},
	}

	proofOfPossession := signer.NewProofOfPossession(h.BlsSecretKey)
	nodeID, err := ids.NodeIDFromString(h.NodeID)
	if err != nil {
		return ids.Empty, err
	}

	unsignedTx, err := wallet.P().Builder().NewAddPermissionlessValidatorTx(
		&txs.SubnetValidator{
			Validator: txs.Validator{
				NodeID: nodeID,
				End:    uint64(time.Now().Add(validator.Duration).Unix()),
				Wght:   validator.StakeAmount,
			},
			Subnet: ids.Empty,
		},
		proofOfPossession,
		wallet.P().Builder().Context().AVAXAssetID,
		owner,
		owner,
		validator.DelegationFee,
	)
	if err != nil {
		return ids.Empty, fmt.Errorf("error building tx: %w", err)
	}

	tx := txs.Tx{Unsigned: unsignedTx}
	if err := wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return ids.Empty, fmt.Errorf("error signing tx: %w", err)
	}

	ctx, cancel := utils.GetAPIContext()
	defer cancel()
	err = wallet.P().IssueTx(
		&tx,
		common.WithContext(ctx),
	)

	if err != nil {
		if ctx.Err() != nil {
			err = fmt.Errorf("timeout issuing/verifying tx with ID %s: %w", tx.ID(), err)
		} else {
			err = fmt.Errorf("error issuing tx with ID %s: %w", tx.ID(), err)
		}
		return ids.Empty, err
	}

	return tx.ID(), nil
}

func GetMinStakingAmount(network avalanche.Network) (uint64, error) {
	pClient := platformvm.NewClient(network.Endpoint)
	ctx, cancel := utils.GetAPIContext()
	defer cancel()
	minValStake, _, err := pClient.GetMinStake(ctx, ids.Empty)
	if err != nil {
		return 0, err
	}
	return minValStake, nil
}

func (h *Node) SetNodeBLSKey(signingKeyPath string) error {
	blsKeyBytes, err := os.ReadFile(signingKeyPath)
	if err != nil {
		return err
	}
	blsSk, err := bls.SecretKeyFromBytes(blsKeyBytes)
	if err != nil {
		return err
	}
	h.BlsSecretKey = blsSk
	return nil
}

func RemoveTmpSDKDir() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("unable to get system user %s", err)
	}
	return os.RemoveAll(filepath.Join(usr.HomeDir, constants.LocalTmpDir))
}

func (h *Node) GetBLSKeyFromRemoteHost() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("unable to get system user %s", err)
	}
	filePath := filepath.Join(constants.CloudNodeStakingPath, constants.BLSKeyFileName)
	localFilePath := filepath.Join(usr.HomeDir, constants.LocalTmpDir, h.NodeID, constants.BLSKeyFileName)
	return h.Download(filePath, localFilePath, constants.SSHFileOpsTimeout)
}

// HandleBLSKey gets BLS information from remote host and sets the BlsSecretKey value in Node object
func (h *Node) HandleBLSKey() error {
	if err := h.GetBLSKeyFromRemoteHost(); err != nil {
		return err
	}
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("unable to get system user %s", err)
	}
	if err := h.SetNodeBLSKey(filepath.Join(usr.HomeDir, constants.LocalTmpDir, h.NodeID, constants.BLSKeyFileName)); err != nil {
		return err
	}
	return RemoveTmpSDKDir()
}
