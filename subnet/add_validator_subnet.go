// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"

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

// AddValidator adds validator to subnet
// Before an Avalanche Node can be added as a validator to a Subnet, the node must already be
// tracking the subnet
// TODO: add more description once node join subnet sdk is done
func (c *Subnet) AddValidator(wallet wallet.Wallet, validatorInput ValidatorParams) (*multisig.Multisig, error) {
	if validatorInput.NodeID == ids.EmptyNodeID {
		return nil, fmt.Errorf(constants.EmptyValidatorNodeIDError)
	}
	if validatorInput.Duration == 0 {
		return nil, fmt.Errorf(constants.EmptyValidatorDurationError)
	}
	if validatorInput.Weight == 0 {
		return nil, fmt.Errorf(constants.EmptyValidatorWeightError)
	}
	if c.SubnetID == ids.Empty {
		return nil, fmt.Errorf(constants.EmptySubnetIDEError)
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
