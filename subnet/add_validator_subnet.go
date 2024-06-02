// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subnet

import (
	"fmt"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/multisig"
	"github.com/ava-labs/avalanche-tooling-sdk-go/wallet"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"golang.org/x/exp/slices"
	"golang.org/x/net/context"
)

type ValidatorParams struct {
	NodeID ids.NodeID

	Duration time.Duration

	Weight uint64

	Network avalanche.Network
}

// AddValidator adds validator to subnet
func (c *Subnet) AddValidator(wallet wallet.Wallet, validatorInput ValidatorParams) (*multisig.Multisig, error) {
	controlKeys, threshold, err := GetOwners(validatorInput.Network, c.SubnetID)
	if err != nil {
		return nil, err
	}
	pChainAddr, err := wallet.Keychain.P()
	if err != nil {
		return nil, err
	}
	var subnetAuthKeysStr []string
	for _, subnetAuthKey := range subnetAuthKeysStr {
		subnetAuthKeysStr = append(subnetAuthKeysStr, subnetAuthKey)
	}
	if err := checkSubnetAuthKeys(pChainAddr, subnetAuthKeysStr, controlKeys, threshold); err != nil {
		return nil, err
	}
	validator := &txs.SubnetValidator{
		Validator: txs.Validator{
			NodeID: validatorInput.NodeID,
			End:    uint64(time.Now().Add(validatorInput.Duration).Unix()),
			Wght:   validatorInput.Weight,
		},
		Subnet: c.SubnetID,
	}

	wallet.SetSubnetAuthMultisig(c.DeployInfo.SubnetAuthKeys)
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

func checkSubnetAuthKeys(walletKeys []string, subnetAuthKeys []string, controlKeys []string, threshold uint32) error {
	for _, walletKey := range walletKeys {
		if slices.Contains(controlKeys, walletKey) && !slices.Contains(subnetAuthKeys, walletKey) {
			return fmt.Errorf("wallet key %s is a subnet control key so it must be included in subnet auth keys", walletKey)
		}
	}
	if len(subnetAuthKeys) != int(threshold) {
		return fmt.Errorf("number of given subnet auth differs from the threshold")
	}
	for _, subnetAuthKey := range subnetAuthKeys {
		found := false
		for _, controlKey := range controlKeys {
			if subnetAuthKey == controlKey {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("subnet auth key %s does not belong to control keys", subnetAuthKey)
		}
	}
	return nil
}

func GetOwners(network avalanche.Network, subnetID ids.ID) ([]string, uint32, error) {
	pClient := platformvm.NewClient(network.Endpoint)
	ctx := context.Background()
	subnetResponse, err := pClient.GetSubnet(ctx, subnetID)
	if err != nil {
		return nil, 0, fmt.Errorf("subnet tx %s query error: %w", subnetID, err)
	}
	controlKeys := subnetResponse.ControlKeys
	threshold := subnetResponse.Threshold
	hrp := network.HRP()
	controlKeysStrs, err := utils.MapE(
		controlKeys,
		func(addr ids.ShortID) (string, error) {
			return address.Format("P", hrp, addr[:])
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return controlKeysStrs, threshold, nil
}
