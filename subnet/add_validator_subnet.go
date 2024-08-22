// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"errors"
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/validator"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"golang.org/x/net/context"
)

var (
	ErrEmptyValidatorNodeID   = errors.New("validator node id is not provided")
	ErrEmptyValidatorDuration = errors.New("validator duration is not provided")
	ErrEmptyValidatorWeight   = errors.New("validator weight is not provided")
	ErrEmptySubnetID          = errors.New("subnet ID is not provided")
)

// AddValidator adds validator to subnet
// Before an Avalanche Node can be added as a validator to a Subnet, the node must already be
// tracking the subnet
// TODO: add more description once node join subnet sdk is done
func (c *Subnet) AddValidator(wallet wallet.Wallet, validatorInput validator.SubnetValidatorParams) (*multisig.Multisig, error) {
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
