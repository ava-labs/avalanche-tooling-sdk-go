// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/staking"
)

// ProvideStakingFiles generates the files needed to validate the primary network:
//   - staker.crt, staker.key, more information can be found at https://docs.avax.network/nodes/validate/how-to-stake#secret-management
//   - The file containing the node's BLS information: signer.key (more information can be found at https://docs.avax.network/cross-chain/avalanche-warp-messaging/deep-dive#bls-multi-signatures-with-public-key-aggregation)
//
// and stores them in the provided directory in argument in local machine
// and subsequently uploads these files into the remote host in /home/ubuntu/.avalanchego/staking/
// directory
func (h *Node) ProvideStakingFiles(keyPath string) error {
	if nodeID, err := GenerateStakingFiles(keyPath); err != nil {
		return err
	} else {
		h.Logger.Infof("Generated Staking Cert and Key for NodeID: %s in folder %s", nodeID.String(), keyPath)
	}
	return h.RunSSHUploadStakingFiles(keyPath)
}

// GenerateStakingFiles generates the following files: staker.crt, staker.key and signer.key
// and stores them in the provided directory in argument in local machine
func GenerateStakingFiles(keyPath string) (ids.NodeID, error) {
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
