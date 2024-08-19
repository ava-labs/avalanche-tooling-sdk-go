// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"os"
	"time"

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

type ValidatorParams struct {
	NodeID ids.NodeID

	Duration time.Duration

	Weight uint64
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

// ValidatePrimaryNetwork adds node as primary network validator.
// It adds the node in the specified network (Fuji / Mainnet / Devnet)
// and uses the wallet provided in the argument to pay for the transaction fee
func (h *Node) ValidatePrimaryNetwork(
	network avalanche.Network,
	validator ValidatorParams,
	wallet wallet.Wallet,
) (ids.ID, error) {
	minValStake, err := GetMinStakingAmount(network)
	if err != nil {
		return ids.Empty, err
	}

	if validator.Weight < minValStake {
		return ids.Empty, fmt.Errorf("invalid weight, must be greater than or equal to %d: %d", minValStake, validator.Weight)
	}

	delegationFee := network.GenesisParams().MinDelegationFee

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
				Wght:   validator.Weight,
			},
			Subnet: ids.Empty,
		},
		proofOfPossession,
		wallet.P().Builder().Context().AVAXAssetID,
		owner,
		owner,
		delegationFee,
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
