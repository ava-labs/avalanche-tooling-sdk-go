// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package validatormanagertypes

type ValidatorManagementType string

const (
	ProofOfStake                 = "Proof Of Stake"
	ProofOfStakeNative           = "Proof Of Stake Native"
	ProofOfStakeERC20            = "Proof Of Stake ERC20"
	ProofOfAuthority             = "Proof Of Authority"
	UndefinedValidatorManagement = "Undefined Validator Management"
)

func ValidatorManagementTypeFromString(s string) ValidatorManagementType {
	switch s {
	case ProofOfStake:
		return ProofOfStake
	case ProofOfStakeNative:
		return ProofOfStakeNative
	case ProofOfStakeERC20:
		return ProofOfStakeERC20
	case ProofOfAuthority:
		return ProofOfAuthority
	default:
		return UndefinedValidatorManagement
	}
}

// IsPoS returns true if the validator management type is any Proof of Stake variant
// This includes ProofOfStake (for backward compatibility), ProofOfStakeNative, and ProofOfStakeERC20
func IsPoS(vmt ValidatorManagementType) bool {
	return vmt == ProofOfStake || vmt == ProofOfStakeNative || vmt == ProofOfStakeERC20
}

// IsPoSNative returns true if the validator management type is Proof of Stake with native token
// This includes both ProofOfStake (for backward compatibility) and ProofOfStakeNative
func IsPoSNative(vmt ValidatorManagementType) bool {
	return vmt == ProofOfStake || vmt == ProofOfStakeNative
}

// IsPoSERC20 returns true if the validator management type is Proof of Stake with ERC20 token
func IsPoSERC20(vmt ValidatorManagementType) bool {
	return vmt == ProofOfStakeERC20
}
