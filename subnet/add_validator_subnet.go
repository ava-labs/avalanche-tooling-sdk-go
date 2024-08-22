// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

var (
	ErrEmptyValidatorNodeID   = errors.New("validator node id is not provided")
	ErrEmptyValidatorDuration = errors.New("validator duration is not provided")
	ErrEmptySubnetID          = errors.New("subnet ID is not provided")
	ErrEmptySubnetAuth        = errors.New("no subnet auth keys is provided")
)

type SubnetValidatorParams struct {
	// NodeID is the unique identifier of the node to be added as a validator on the specified Subnet.
	NodeID ids.NodeID
	// Duration is how long the node will be staking the Subnet
	// Duration has to be less than or equal to the duration that the node will be validating the Primary
	// Network
	Duration time.Duration
	// Weight is the validator's weight when sampling validators.
	// Weight for subnet validators is set to 20 by default
	Weight uint64
}

// AddValidator adds validator to subnet
// Before an Avalanche Node can be added as a validator to a Subnet, the node must already be
// tracking the subnet, which can be done by calling SyncSubnets in node package
func (c *Subnet) AddValidator(wallet wallet.Wallet, validatorInput SubnetValidatorParams) (*multisig.Multisig, error) {
	if validatorInput.NodeID == ids.EmptyNodeID {
		return nil, ErrEmptyValidatorNodeID
	}
	if validatorInput.Duration == 0 {
		return nil, ErrEmptyValidatorDuration
	}
	if validatorInput.Weight == 0 {
		validatorInput.Weight = 20
	}
	if c.SubnetID == ids.Empty {
		return nil, ErrEmptySubnetID
	}
	if len(c.DeployInfo.SubnetAuthKeys) == 0 {
		return nil, ErrEmptySubnetAuth
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
