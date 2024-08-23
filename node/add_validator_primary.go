// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"

	remoteconfig "github.com/ava-labs/avalanche-tooling-sdk-go/node/config"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/subnet"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"golang.org/x/net/context"
)

// ValidatePrimaryNetwork adds node as primary network validator.
// It adds the node in the specified network (Fuji / Mainnet / Devnet)
// and uses the wallet provided in the argument to pay for the transaction fee
func (h *Node) ValidatePrimaryNetwork(
	network avalanche.Network,
	validatorParams validator.PrimaryNetworkValidatorParams,
	wallet wallet.Wallet,
) (ids.ID, error) {
	if validatorParams.NodeID == ids.EmptyNodeID {
		return ids.Empty, subnet.ErrEmptyValidatorNodeID
	}

	if validatorParams.Duration == 0 {
		return ids.Empty, subnet.ErrEmptyValidatorDuration
	}

	minValStake, err := network.GetMinStakingAmount()
	if err != nil {
		return ids.Empty, err
	}

	if validatorParams.StakeAmount < minValStake {
		return ids.Empty, fmt.Errorf("invalid weight, must be greater than or equal to %d: %d", minValStake, validatorParams.StakeAmount)
	}

	if validatorParams.DelegationFee == 0 {
		validatorParams.DelegationFee = network.GenesisParams().MinDelegationFee
	}

	if err = h.GetBLSKeyFromRemoteHost(); err != nil {
		return ids.Empty, fmt.Errorf("unable to set BLS key of node from remote host due to %w", err)
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
				End:    uint64(time.Now().Add(validatorParams.Duration).Unix()),
				Wght:   validatorParams.StakeAmount,
			},
			Subnet: ids.Empty,
		},
		proofOfPossession,
		wallet.P().Builder().Context().AVAXAssetID,
		owner,
		owner,
		validatorParams.DelegationFee,
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

// GetBLSKeyFromRemoteHost gets BLS information from remote host and sets the BlsSecretKey value in Node object
func (h *Node) GetBLSKeyFromRemoteHost() error {
	blsKeyBytes, err := h.ReadFileBytes(remoteconfig.GetRemoteBLSKeyFile(), constants.SSHFileOpsTimeout)
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
