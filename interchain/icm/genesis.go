// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package icm

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/subnet-evm/core"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

const (
	// RegistryContractAddressAtGenesis is the canonical address for the TeleporterRegistry
	// contract when pre-allocated in genesis.
	RegistryContractAddressAtGenesis = "0xF86Cb19Ad8405AEFa7d09C778215D2Cb6eBfB228"

	messengerVersionAtGenesis = "0x1"
)

//go:embed smart_contracts/deployed_messenger_bytecode_v1.0.0.txt
var deployedMessengerBytecode []byte

//go:embed smart_contracts/deployed_registry_bytecode_v1.0.0.txt
var deployedRegistryBytecode []byte

func setSimpleStorageValue(
	storage map[common.Hash]common.Hash,
	slot string,
	value string,
) {
	storage[common.HexToHash(slot)] = common.HexToHash(value)
}

func hexFill32(s string) string {
	return fmt.Sprintf("%064s", utils.TrimHexa(s))
}

func setMappingStorageValue(
	storage map[common.Hash]common.Hash,
	slot string,
	key string,
	value string,
) error {
	slot = hexFill32(slot)
	key = hexFill32(key)
	storageKey := key + slot
	storageKeyBytes, err := hex.DecodeString(storageKey)
	if err != nil {
		return err
	}
	storage[crypto.Keccak256Hash(storageKeyBytes)] = common.HexToHash(value)
	return nil
}

// AddMessengerContractToAllocations adds the TeleporterMessenger contract to genesis allocations.
// This pre-deploys the messenger contract at its canonical address with initialized storage.
func AddMessengerContractToAllocations(
	allocs core.GenesisAlloc,
) {
	const (
		blockchainIDSlot = "0x0"
		messageNonceSlot = "0x1"
	)
	storage := map[common.Hash]common.Hash{}
	setSimpleStorageValue(storage, blockchainIDSlot, "0x1")
	setSimpleStorageValue(storage, messageNonceSlot, "0x1")
	deployedMessengerBytes := common.FromHex(strings.TrimSpace(string(deployedMessengerBytecode)))
	allocs[common.HexToAddress(DefaultMessengerContractAddress)] = core.GenesisAccount{
		Balance: big.NewInt(0),
		Code:    deployedMessengerBytes,
		Storage: storage,
		Nonce:   1,
	}
	allocs[common.HexToAddress(DefaultMessengerDeployerAddress)] = core.GenesisAccount{
		Balance: big.NewInt(0),
		Nonce:   1,
	}
}

// AddRegistryContractToAllocations adds the TeleporterRegistry contract to genesis allocations.
// The registry is initialized with the messenger contract registered at version 1.
func AddRegistryContractToAllocations(
	allocs core.GenesisAlloc,
) error {
	const (
		latestVersionSlot    = "0x0"
		versionToAddressSlot = "0x1"
		addressToVersionSlot = "0x2"
	)
	storage := map[common.Hash]common.Hash{}
	setSimpleStorageValue(storage, latestVersionSlot, messengerVersionAtGenesis)
	if err := setMappingStorageValue(storage, versionToAddressSlot, messengerVersionAtGenesis, DefaultMessengerContractAddress); err != nil {
		return err
	}
	if err := setMappingStorageValue(storage, addressToVersionSlot, DefaultMessengerContractAddress, messengerVersionAtGenesis); err != nil {
		return err
	}
	deployedRegistryBytes := common.FromHex(strings.TrimSpace(string(deployedRegistryBytecode)))
	allocs[common.HexToAddress(RegistryContractAddressAtGenesis)] = core.GenesisAccount{
		Balance: big.NewInt(0),
		Code:    deployedRegistryBytes,
		Storage: storage,
		Nonce:   1,
	}
	return nil
}
