// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package ledger

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/platformvm"

	"github.com/ava-labs/avalanche-tooling-sdk-go/keychain/ledger"
	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

const (
	maxIndexToSearch           = 1000
	maxIndexToSearchForBalance = 100
)

type LedgerDevice struct {
	ledger.Ledger
}

func New() (*LedgerDevice, error) {
	device, err := ledger.New()
	if err != nil {
		return nil, err
	}
	dev := LedgerDevice{
		Ledger: device,
	}
	return &dev, nil
}

func (dev *LedgerDevice) FindAddresses(addresses []string, maxIndex uint32) (map[string]uint32, error) {
	addressesIDs, err := address.ParseToIDs(addresses)
	if err != nil {
		return nil, fmt.Errorf("failure parsing ledger addresses: %w", err)
	}
	// for all ledger indices to search for, find if the ledger address belongs to the input
	// addresses and, if so, add an index association to indexMap.
	// breaks the loop if all addresses were found
	if maxIndex == 0 {
		maxIndex = maxIndexToSearch
	}
	indices := map[string]uint32{}
	for index := uint32(0); index < maxIndex; index++ {
		pubKeys, err := dev.PubKeys([]uint32{index})
		if err != nil {
			return nil, err
		}
		ledgerAddress := pubKeys[0].Address()
		for addressIndex, addr := range addressesIDs {
			if addr == ledgerAddress {
				indices[addresses[addressIndex]] = index
			}
		}
		if len(indices) == len(addresses) {
			break
		}
	}
	return indices, nil
}

// FindFunds searches for a set of indices that pay a given amount
func (dev *LedgerDevice) FindFunds(
	network network.Network,
	amount uint64,
	maxIndex uint32,
) ([]uint32, error) {
	pClient := platformvm.NewClient(network.Endpoint)
	totalBalance := uint64(0)
	indices := []uint32{}
	if maxIndex == 0 {
		maxIndex = maxIndexToSearchForBalance
	}
	for index := uint32(0); index < maxIndex; index++ {
		pubKeys, err := dev.PubKeys([]uint32{index})
		if err != nil {
			return []uint32{}, err
		}
		ledgerAddress := pubKeys[0].Address()
		ctx, cancel := utils.GetAPIContext()
		resp, err := pClient.GetBalance(ctx, []ids.ShortID{ledgerAddress})
		cancel()
		if err != nil {
			return nil, err
		}
		if resp.Balance > 0 {
			totalBalance += uint64(resp.Balance)
			indices = append(indices, index)
		}
		if totalBalance >= amount {
			break
		}
	}
	if totalBalance < amount {
		return nil, fmt.Errorf("not enough funds on ledger")
	}
	return indices, nil
}
