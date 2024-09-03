// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package evm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/ava-labs/subnet-evm/accounts/abi/bind"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ethereum/go-ethereum/common"
)

var ErrFailedReceiptStatus = fmt.Errorf("failed receipt status")

type SignatureKind int64

const (
	Constructor SignatureKind = iota
	Method
	Event
)

type PaymentKind int64

const (
	View PaymentKind = iota
	Payable
	NonPayable
)

// splitTypes splits a string of comma separated type esps into a slice of type esps:
// - considering a list of types esps surrounded by (nested) parenthesis as one struct type esp
// - considering a type esp surrounded by brackets as one array type esp
// note: it just parses the first level of type esps: parsing of nested subtypes is called recursively by getABIMaps
//
// ie:
// "bool,int" maps to ["bool", "int"]  - 2 primitive types bool and int
// "(bool,int)" maps to ["(bool,int)"] - 1 struct type (bool,int)
// "(bool,int),bool" maps to ["(bool,int)","bool"] - 1 struct type (bool,int) + 1 primitive type bool
// "[(bool,int)],[bool]" maps to ["[(bool,int)]","[bool]"] - 1 array of structs (bool,int) type + 1 array of bools
//
// TODO: manage all recursion here, returning a list of types, not strings, where compound types are supported
// (so, returning a tree)
func splitTypes(s string) []string {
	words := []string{}
	word := ""
	parenthesisCount := 0
	insideBrackets := false
	for _, rune := range s {
		c := string(rune)
		if parenthesisCount > 0 {
			word += c
			if c == "(" {
				parenthesisCount++
			}
			if c == ")" {
				parenthesisCount--
				if parenthesisCount == 0 {
					words = append(words, word)
					word = ""
				}
			}
			continue
		}
		if insideBrackets {
			word += c
			if c == "]" {
				words = append(words, word)
				word = ""
				insideBrackets = false
			}
			continue
		}
		if c == " " || c == "," || c == "(" || c == "[" {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		}
		if c == " " || c == "," {
			continue
		}
		if c == "(" {
			parenthesisCount++
		}
		if c == "[" {
			insideBrackets = true
		}
		word += c
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// for a list of strings that specifie types, it generates
// a list of ethereum ABI descriptions,
// compatible with the subnet-evm bind library
//
// note: as, for structs, bind calls check ABI field names against golang struct field
// names, input [values] are passed, so as to get appropriate ABI names
func getABIMaps(
	types []string,
	values interface{},
) ([]map[string]interface{}, error) {
	r := []map[string]interface{}{}
	for i, t := range types {
		var (
			value      interface{}
			name       string
			structName string
		)
		rt := reflect.ValueOf(values)
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if rt.Kind() == reflect.Slice {
			if rt.Len() != len(types) {
				if rt.Len() == 1 {
					return getABIMaps(types, rt.Index(0).Interface())
				} else {
					return nil, fmt.Errorf(
						"inconsistency in slice len between method esp %q and given values %#v: expected %d got %d",
						types,
						values,
						len(types),
						rt.Len(),
					)
				}
			}
			value = rt.Index(i).Interface()
		} else if rt.Kind() == reflect.Struct {
			if rt.NumField() < len(types) {
				return nil, fmt.Errorf(
					"inconsistency in struct len between method esp %q and given values %#v: expected %d got %d",
					types,
					values,
					len(types),
					rt.NumField(),
				)
			}
			name = rt.Type().Field(i).Name
			structName = rt.Type().Field(i).Type.Name()
			value = rt.Field(i).Interface()
		}
		m := map[string]interface{}{}
		switch {
		case string(t[0]) == "(":
			// struct type
			var err error
			t, err = utils.RemoveSurrounding(t, "(", ")")
			if err != nil {
				return nil, err
			}
			m["components"], err = getABIMaps(splitTypes(t), value)
			if err != nil {
				return nil, err
			}
			if structName != "" {
				m["internalType"] = "struct " + structName
			} else {
				m["internalType"] = "tuple"
			}
			m["type"] = "tuple"
			m["name"] = name
		case string(t[0]) == "[":
			var err error
			t, err = utils.RemoveSurrounding(t, "[", "]")
			if err != nil {
				return nil, err
			}
			if string(t[0]) == "(" {
				t, err = utils.RemoveSurrounding(t, "(", ")")
				if err != nil {
					return nil, err
				}
				rt := reflect.ValueOf(value)
				if rt.Kind() != reflect.Slice {
					return nil, fmt.Errorf("expected value for field %d of esp %q to be an slice", i, types)
				}
				value = reflect.Zero(rt.Type().Elem()).Interface()
				structName = rt.Type().Elem().Name()
				m["components"], err = getABIMaps(splitTypes(t), value)
				if err != nil {
					return nil, err
				}
				if structName != "" {
					m["internalType"] = "struct " + structName + "[]"
				} else {
					m["internalType"] = "tuple[]"
				}
				m["type"] = "tuple[]"
				m["name"] = name
			} else {
				m["internalType"] = fmt.Sprintf("%s[]", t)
				m["type"] = fmt.Sprintf("%s[]", t)
				m["name"] = name
			}
		default:
			m["internalType"] = t
			m["type"] = t
			m["name"] = name
		}
		r = append(r, m)
	}
	return r, nil
}

// ParseMethodSignature parses method/event [signature]
// of format "name(inputs)->(outputs)", where:
//   - name is optional
//   - ->(outputs) is optional
//   - inputs and outputs are a comma separated list of type esps that follow the
//     format of splitTypes
//
// generates a ethereum ABI especification
// that can be used in the subnet-evm bind library.
//
// note: as, for structs, bind calls check ABI field names against golang struct field
// names, input [values] are passed, so as to get appropriate ABI names
func ParseMethodSignature(
	signature string,
	kind SignatureKind,
	indexedFields []int,
	paymentKind PaymentKind,
	values ...interface{},
) (string, string, error) {
	inputsOutputsIndex := strings.Index(signature, "(")
	if inputsOutputsIndex == -1 {
		return signature, "", nil
	}
	name := signature[:inputsOutputsIndex]
	typesSignature := signature[inputsOutputsIndex:]
	inputTypesSignature := ""
	outputTypesSignature := ""
	arrowIndex := strings.Index(typesSignature, "->")
	if arrowIndex == -1 {
		inputTypesSignature = typesSignature
	} else {
		inputTypesSignature = typesSignature[:arrowIndex]
		outputTypesSignature = typesSignature[arrowIndex+2:]
	}
	var err error
	inputTypesSignature, err = utils.RemoveSurrounding(inputTypesSignature, "(", ")")
	if err != nil {
		return "", "", err
	}
	outputTypesSignature, err = utils.RemoveSurrounding(outputTypesSignature, "(", ")")
	if err != nil {
		return "", "", err
	}
	inputTypes := splitTypes(inputTypesSignature)
	outputTypes := splitTypes(outputTypesSignature)
	inputsMaps, err := getABIMaps(inputTypes, values)
	if err != nil {
		return "", "", err
	}
	outputsMaps, err := getABIMaps(outputTypes, nil)
	if err != nil {
		return "", "", err
	}
	abiMap := map[string]interface{}{
		"inputs": inputsMaps,
	}
	switch kind {
	case Constructor:
		abiMap["type"] = "constructor"
		abiMap["stateMutability"] = "nonpayable"
	case Method:
		abiMap["type"] = "function"
		abiMap["name"] = name
		abiMap["outputs"] = outputsMaps
		switch paymentKind {
		case Payable:
			abiMap["stateMutability"] = "payable"
		case View:
			abiMap["stateMutability"] = "view"
		case NonPayable:
			abiMap["stateMutability"] = "nonpayable"
		default:
			return "", "", fmt.Errorf("unsupported payment kind %d", paymentKind)
		}
	case Event:
		abiMap["type"] = "event"
		abiMap["name"] = name
		for i := range inputsMaps {
			if utils.Belongs(indexedFields, i) {
				inputsMaps[i]["indexed"] = true
			}
		}
	default:
		return "", "", fmt.Errorf("unsupported signature kind %d", kind)
	}
	abiMap["inputs"] = inputsMaps
	abiSlice := []map[string]interface{}{abiMap}
	abiBytes, err := json.MarshalIndent(abiSlice, "", "  ")
	if err != nil {
		return "", "", err
	}
	return name, string(abiBytes), nil
}

func TxToMethod(
	rpcURL string,
	privateKey string,
	contractAddress common.Address,
	payment *big.Int,
	methodSignature string,
	params ...interface{},
) (*types.Transaction, *types.Receipt, error) {
	paymentKind := NonPayable
	if payment != nil {
		paymentKind = Payable
	}
	methodName, methodABI, err := ParseMethodSignature(methodSignature, Method, nil, paymentKind, params...)
	if err != nil {
		return nil, nil, err
	}
	metadata := &bind.MetaData{
		ABI: methodABI,
	}
	abi, err := metadata.GetAbi()
	if err != nil {
		return nil, nil, err
	}
	client, err := GetClient(rpcURL)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()
	contract := bind.NewBoundContract(contractAddress, *abi, client, client, client)
	txOpts, err := GetTxOptsWithSigner(client, privateKey)
	if err != nil {
		return nil, nil, err
	}
	txOpts.Value = payment
	tx, err := contract.Transact(txOpts, methodName, params...)
	if err != nil {
		return nil, nil, err
	}
	receipt, success, err := WaitForTransaction(client, tx)
	if err != nil {
		return tx, nil, err
	} else if !success {
		return tx, receipt, ErrFailedReceiptStatus
	}
	return tx, receipt, nil
}

func CallToMethod(
	rpcURL string,
	contractAddress common.Address,
	methodEsp string,
	params ...interface{},
) ([]interface{}, error) {
	methodName, methodABI, err := ParseMethodSignature(methodEsp, Method, nil, View, params...)
	if err != nil {
		return nil, err
	}
	metadata := &bind.MetaData{
		ABI: methodABI,
	}
	abi, err := metadata.GetAbi()
	if err != nil {
		return nil, err
	}
	client, err := GetClient(rpcURL)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	contract := bind.NewBoundContract(contractAddress, *abi, client, client, client)
	var out []interface{}
	err = contract.Call(&bind.CallOpts{}, &out, methodName, params...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func DeployContract(
	rpcURL string,
	privateKey string,
	binBytes []byte,
	methodEsp string,
	params ...interface{},
) (common.Address, error) {
	_, methodABI, err := ParseMethodSignature(methodEsp, Constructor, nil, NonPayable, params...)
	if err != nil {
		return common.Address{}, err
	}
	metadata := &bind.MetaData{
		ABI: methodABI,
		Bin: string(binBytes),
	}
	abi, err := metadata.GetAbi()
	if err != nil {
		return common.Address{}, err
	}
	bin := common.FromHex(metadata.Bin)
	client, err := GetClient(rpcURL)
	if err != nil {
		return common.Address{}, err
	}
	defer client.Close()
	txOpts, err := GetTxOptsWithSigner(client, privateKey)
	if err != nil {
		return common.Address{}, err
	}
	address, tx, _, err := bind.DeployContract(txOpts, *abi, bin, client, params...)
	if err != nil {
		return common.Address{}, err
	}
	if _, success, err := WaitForTransaction(client, tx); err != nil {
		return common.Address{}, err
	} else if !success {
		return common.Address{}, ErrFailedReceiptStatus
	}
	return address, nil
}

func UnpackLog(
	eventEsp string,
	indexedFields []int,
	log types.Log,
	event interface{},
) error {
	eventName, eventABI, err := ParseMethodSignature(eventEsp, Event, indexedFields, NonPayable, event)
	if err != nil {
		return err
	}
	metadata := &bind.MetaData{
		ABI: eventABI,
	}
	abi, err := metadata.GetAbi()
	if err != nil {
		return err
	}
	contract := bind.NewBoundContract(common.Address{}, *abi, nil, nil, nil)
	return contract.UnpackLog(event, eventName, log)
}
