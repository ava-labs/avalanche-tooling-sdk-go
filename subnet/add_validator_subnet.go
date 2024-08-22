// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"errors"
	"time"

	"github.com/ava-labs/avalanchego/ids"
)

var (
	ErrEmptyValidatorNodeID   = errors.New("validator node id is not provided")
	ErrEmptyValidatorDuration = errors.New("validator duration is not provided")
	ErrEmptySubnetID          = errors.New("subnet ID is not provided")
	ErrEmptySubnetAuth        = errors.New("no subnet auth keys is provided")
)

type ValidatorParams struct {
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
