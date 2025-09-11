// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blockchain

import (
	"encoding/json"

	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/signer"

	"github.com/ava-labs/avalanche-tooling-sdk-go/network"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"

	"github.com/ava-labs/avalanchego/ids"
)

func GetSubnet(subnetID ids.ID, network network.Network) (platformvm.GetSubnetClientResponse, error) {
	api := network.Endpoint
	pClient := platformvm.NewClient(api)
	ctx, cancel := utils.GetAPIContext()
	defer cancel()
	return pClient.GetSubnet(ctx, subnetID)
}

func ConvertToBLSProofOfPossession(publicKey, proofOfPossesion string) (signer.ProofOfPossession, error) {
	type jsonProofOfPossession struct {
		PublicKey         string
		ProofOfPossession string
	}
	jsonPop := jsonProofOfPossession{
		PublicKey:         publicKey,
		ProofOfPossession: proofOfPossesion,
	}
	popBytes, err := json.Marshal(jsonPop)
	if err != nil {
		return signer.ProofOfPossession{}, err
	}
	pop := &signer.ProofOfPossession{}
	err = pop.UnmarshalJSON(popBytes)
	if err != nil {
		return signer.ProofOfPossession{}, err
	}
	return *pop, nil
}
