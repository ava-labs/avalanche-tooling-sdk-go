// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"errors"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"
	"os"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
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

var (
	ErrEmptyValidatorNodeID   = errors.New("validator node id is not provided")
	ErrEmptyValidatorDuration = errors.New("validator duration is not provided")
	ErrEmptyValidatorWeight   = errors.New("validator weight is not provided")
	ErrEmptySubnetID          = errors.New("subnet ID is not provided")
)

// addNodeAsPrimaryNetworkValidator returns bool if node is added as primary network validator
// as it impacts the output in adding node as subnet validator in the next steps
func addNodeAsPrimaryNetworkValidator(
	deployer *subnet.PublicDeployer,
	network models.Network,
	kc *keychain.Keychain,
	nodeID ids.NodeID,
	nodeIndex int,
	instanceID string,
) error {
	if isValidator, err := checkNodeIsPrimaryNetworkValidator(nodeID, network); err != nil {
		return err
	} else if !isValidator {
		signingKeyPath := app.GetNodeBLSSecretKeyPath(instanceID)
		return joinAsPrimaryNetworkValidator(deployer, network, kc, nodeID, nodeIndex, signingKeyPath, true)
	}
	return nil
}

func joinAsPrimaryNetworkValidator(
	deployer *subnet.PublicDeployer,
	network models.Network,
	kc *keychain.Keychain,
	nodeID ids.NodeID,
	nodeIndex int,
	signingKeyPath string,
	nodeCmd bool,
) error {
	var (
		start time.Time
		err   error
	)
	minValStake, err := GetMinStakingAmount(network)
	if err != nil {
		return err
	}
	if weight == 0 {
		weight, err = PromptWeightPrimaryNetwork(network)
		if err != nil {
			return err
		}
	}
	if weight < minValStake {
		return fmt.Errorf("invalid weight, must be greater than or equal to %d: %d", minValStake, weight)
	}
	start, duration, err = GetTimeParametersPrimaryNetwork(network, nodeIndex, duration, startTimeStr, nodeCmd)
	if err != nil {
		return err
	}

	recipientAddr := kc.Addresses().List()[0]
	// we set the starting time for node to be a Primary Network Validator to be in 1 minute
	// we use min delegation fee as default
	delegationFee := network.GenesisParams().MinDelegationFee
	blsKeyBytes, err := os.ReadFile(signingKeyPath)
	if err != nil {
		return err
	}
	blsSk, err := bls.SecretKeyFromBytes(blsKeyBytes)
	if err != nil {
		return err
	}
	if _, err := deployer.AddPermissionlessValidator(
		ids.Empty,
		ids.Empty,
		nodeID,
		weight,
		uint64(start.Unix()),
		uint64(start.Add(duration).Unix()),
		recipientAddr,
		delegationFee,
		nil,
		signer.NewProofOfPossession(blsSk),
	); err != nil {
		return err
	}
	return nil
}

func (d *PublicDeployer) AddPermissionlessValidator(
	subnetID ids.ID,
	subnetAssetID ids.ID,
	nodeID ids.NodeID,
	stakeAmount uint64,
	startTime uint64,
	endTime uint64,
	recipientAddr ids.ShortID,
	delegationFee uint32,
	popBytes []byte,
	proofOfPossession *signer.ProofOfPossession,
) (ids.ID, error) {
	wallet, err := d.loadWallet(subnetID)
	if err != nil {
		return ids.Empty, err
	}
	if subnetAssetID == ids.Empty {
		subnetAssetID = wallet.P().Builder().Context().AVAXAssetID
	}
	// popBytes is a marshalled json object containing publicKey and proofOfPossession of the node's BLS info
	txID, err := d.issueAddPermissionlessValidatorTX(recipientAddr, stakeAmount, subnetID, nodeID, subnetAssetID, startTime, endTime, wallet, delegationFee, popBytes, proofOfPossession)
	if err != nil {
		return ids.Empty, err
	}
	return txID, nil
}

func (c *Subnet) issueAddPermissionlessValidatorTX(
	recipientAddr ids.ShortID,
	stakeAmount uint64,
	subnetID ids.ID,
	nodeID ids.NodeID,
	assetID ids.ID,
	startTime uint64,
	endTime uint64,
	wallet primary.Wallet,
	delegationFee uint32,
	popBytes []byte,
	blsProof *signer.ProofOfPossession,
) (ids.ID, error) {
	options := d.getMultisigTxOptions([]ids.ShortID{})
	owner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs: []ids.ShortID{
			recipientAddr,
		},
	}
	var proofOfPossession signer.Signer
	if subnetID == ids.Empty {
		if popBytes != nil {
			pop := &signer.ProofOfPossession{}
			err := pop.UnmarshalJSON(popBytes)
			if err != nil {
				return ids.Empty, err
			}
			proofOfPossession = pop
		} else {
			proofOfPossession = blsProof
		}
	} else {
		proofOfPossession = &signer.Empty{}
	}

	unsignedTx, err := wallet.P().Builder().NewAddPermissionlessValidatorTx(
		&txs.SubnetValidator{
			Validator: txs.Validator{
				NodeID: nodeID,
				Start:  startTime,
				End:    endTime,
				Wght:   stakeAmount,
			},
			Subnet: subnetID,
		},
		proofOfPossession,
		assetID,
		owner,
		owner,
		delegationFee,
		options...,
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

// AddValidator adds validator to subnet
// Before an Avalanche Node can be added as a validator to a Subnet, the node must already be
// tracking the subnet
// TODO: add more description once node join subnet sdk is done
func (c *Subnet) AddValidator(wallet wallet.Wallet, validatorInput ValidatorParams) (*multisig.Multisig, error) {
	if validatorInput.NodeID == ids.EmptyNodeID {
		return nil, ErrEmptyValidatorNodeID
	}
	if validatorInput.Duration == 0 {
		return nil, ErrEmptyValidatorDuration
	}
	if validatorInput.Weight == 0 {
		return nil, ErrEmptyValidatorWeight
	}
	if c.SubnetID == ids.Empty {
		return nil, ErrEmptySubnetID
	}

	wallet.SetSubnetAuthMultisig(c.DeployInfo.SubnetAuthKeys)

	validator := &txs.SubnetValidator{
		Validator: txs.Validator{
			NodeID: validatorInput.NodeID,
			End:    uint64(time.Now().Add(validatorInput.Duration).Unix()),
			Wght:   validatorInput.Weight,
		},
		Subnet: c.SubnetID,
	}

	unsignedTx, err := wallet.P().Builder().NewAddSubnetValidatorTx(validator)
	if err != nil {
		return nil, fmt.Errorf("error building tx: %w", err)
	}
	tx := txs.Tx{Unsigned: unsignedTx}
	if err := wallet.P().Signer().Sign(context.Background(), &tx); err != nil {
		return nil, fmt.Errorf("error signing tx: %w", err)
	}
	return multisig.New(&tx), nil
}
