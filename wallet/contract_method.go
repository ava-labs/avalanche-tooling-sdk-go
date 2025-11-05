// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

// ContractMethod encapsulates a contract method signature and its parameters
type ContractMethod struct {
	Spec   string        // Method signature (e.g., "transfer(address,uint256)")
	Params []interface{} // Method parameters
}

// NewContractMethod creates a new ContractMethod with the given signature and parameters
func NewContractMethod(spec string, params ...interface{}) ContractMethod {
	return ContractMethod{
		Spec:   spec,
		Params: params,
	}
}
