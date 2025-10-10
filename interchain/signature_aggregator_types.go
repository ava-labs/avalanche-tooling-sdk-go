// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package interchain

// AggregateSignatureRequest is the request structure for the signature aggregator API.
// This is a local copy to avoid dependency issues with icm-services.
type AggregateSignatureRequest struct {
	Message          string `json:"message"`
	SigningSubnetID  string `json:"signingSubnetID"`
	QuorumPercentage uint64 `json:"quorumPercentage"`
	Justification    string `json:"justification"`
}

// AggregateSignatureResponse is the response structure from the signature aggregator API.
// This is a local copy to avoid dependency issues with icm-services.
type AggregateSignatureResponse struct {
	SignedMessage string `json:"signedMessage"`
}
