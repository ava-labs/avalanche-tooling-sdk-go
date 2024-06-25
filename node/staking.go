// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package host

import (
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/staking"
)

func (h *Host) ProvideStakingCertAndKey(keyPath string) error {
	if nodeID, err := h.GenerateNodeCertAndKeys(keyPath); err != nil {
		return err
	} else {
		h.Logger.Infof("Generated Staking Cert and Key for NodeID: %s in folder %s", nodeID.String(), keyPath)
	}
	return h.RunSSHUploadStakingFiles(keyPath)
}

// GenerateNodeCertAndKeys generates a node certificate and keys and return nodeID
func (h *Host) GenerateNodeCertAndKeys(keyPath string) (ids.NodeID, error) {
	if err := os.MkdirAll(keyPath, constants.DefaultPerms755); err != nil {
		return ids.EmptyNodeID, err
	}
	stakerCertFilePath := filepath.Join(keyPath, constants.StakerCertFileName)
	stakerKeyFilePath := filepath.Join(keyPath, constants.StakerKeyFileName)
	blsKeyFilePath := filepath.Join(keyPath, constants.BLSKeyFileName)

	certBytes, keyBytes, err := staking.NewCertAndKeyBytes()
	if err != nil {
		return ids.EmptyNodeID, err
	}
	nodeID, err := utils.ToNodeID(certBytes)
	if err != nil {
		return ids.EmptyNodeID, err
	}
	if err := os.MkdirAll(filepath.Dir(stakerCertFilePath), constants.DefaultPerms755); err != nil {
		return ids.EmptyNodeID, err
	}
	if err := os.WriteFile(stakerCertFilePath, certBytes, constants.WriteReadUserOnlyPerms); err != nil {
		return ids.EmptyNodeID, err
	}
	if err := os.MkdirAll(filepath.Dir(stakerKeyFilePath), constants.DefaultPerms755); err != nil {
		return ids.EmptyNodeID, err
	}
	if err := os.WriteFile(stakerKeyFilePath, keyBytes, constants.WriteReadUserOnlyPerms); err != nil {
		return ids.EmptyNodeID, err
	}
	blsSignerKeyBytes, err := utils.NewBlsSecretKeyBytes()
	if err != nil {
		return ids.EmptyNodeID, err
	}
	if err := os.MkdirAll(filepath.Dir(blsKeyFilePath), constants.DefaultPerms755); err != nil {
		return ids.EmptyNodeID, err
	}
	if err := os.WriteFile(blsKeyFilePath, blsSignerKeyBytes, constants.WriteReadUserOnlyPerms); err != nil {
		return ids.EmptyNodeID, err
	}
	return nodeID, nil
}
