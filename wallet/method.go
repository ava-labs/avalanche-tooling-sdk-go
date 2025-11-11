// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package wallet

import "strings"

// ContractMethod encapsulates a contract method signature and its parameters
type ContractMethod struct {
	Spec   string        // Method signature (e.g., "transfer(address,uint256)")
	Params []interface{} // Method parameters
}

// Method creates a new ContractMethod with the given signature and parameters
func Method(spec string, params ...interface{}) ContractMethod {
	return ContractMethod{
		Spec:   spec,
		Params: params,
	}
}

// Name extracts the method name from the spec
// Examples:
//
//	"transfer(address,uint256)" -> "transfer"
//	"balanceOf(address)->(uint256)" -> "balanceOf"
//	"(string,address,uint256)" -> "(constructor)"
func (m ContractMethod) Name() string {
	spec := strings.TrimSpace(m.Spec)
	if strings.HasPrefix(spec, "(") {
		return "(constructor)"
	}
	if idx := strings.Index(spec, "("); idx > 0 {
		return spec[:idx]
	}
	return spec
}
