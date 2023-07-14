//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jssupport

import (
	"errors"
	"fmt"
	"syscall/js"

	interoperror "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/errors"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

type NamedValue struct {
	Name  string
	Value js.Value
}

func NewNamedValue(name string, val js.Value) NamedValue {
	return NamedValue{
		Name:  name,
		Value: val,
	}
}

// Function with named arguments in JS will look like this someFunc({param1: "value"})
func GetNamedArgument(args []js.Value, name string) (NamedValue, error) {
	if len(args) > 1 {
		return NewNamedValue(name, js.Null()), walleterror.NewValidationError(
			interoperror.Module,
			interoperror.InvalidArgumentsCode,
			interoperror.InvalidArgumentsError,
			errors.New("function with named parameters should receive only one or zero arguments"))
	}

	if len(args) == 1 {
		if args[0].Type() != js.TypeObject {
			return NewNamedValue(name, js.Null()), walleterror.NewValidationError(
				interoperror.Module,
				interoperror.InvalidArgumentsCode,
				interoperror.InvalidArgumentsError,
				fmt.Errorf("arguments should be pack into object, but get value with type %s", args[0].Type().String()))
		}

		val := args[0].Get(name)

		if !val.IsNull() && !val.IsUndefined() {
			return NewNamedValue(name, val), nil
		}
	}

	return NewNamedValue(name, js.Null()),
		walleterror.NewValidationError(
			interoperror.Module,
			interoperror.InvalidArgumentsCode,
			interoperror.InvalidArgumentsError,
			fmt.Errorf("named parameter %s is required", name))
}

func GetOptionalNamedArgument(args []js.Value, name string) (NamedValue, error) {
	if len(args) > 1 {
		return NewNamedValue(name, js.Null()), walleterror.NewValidationError(
			interoperror.Module,
			interoperror.InvalidArgumentsCode,
			interoperror.InvalidArgumentsError,
			errors.New("function with named parameters should receive only one or zero arguments"))
	}

	if len(args) == 1 {
		val := args[0].Get(name)

		if !val.IsNull() && !val.IsUndefined() {
			return NewNamedValue(name, val), nil
		}
	}

	return NewNamedValue(name, js.Null()), nil
}

func GetNamedProperty(object js.Value, name string) (NamedValue, error) {
	val := object.Get(name)

	if !val.IsNull() && !val.IsUndefined() {
		return NewNamedValue(name, val), nil
	}

	return NewNamedValue(name, js.Null()),
		walleterror.NewValidationError(
			interoperror.Module,
			interoperror.MissedRequiredPropertyCode,
			interoperror.MissedRequiredPropertyError,
			fmt.Errorf("property %q is required", name))
}

func EnsureString(arg NamedValue, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if arg.Value.Type() == js.TypeNull {
		return "", nil
	}

	if arg.Value.Type() != js.TypeString {
		return "", fmt.Errorf("argument %q should have type string", arg.Name)
	}

	return arg.Value.String(), nil
}

func EnsureStringArray(arg NamedValue, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}

	if arg.Value.Type() == js.TypeNull {
		return nil, nil
	}

	if arg.Value.Length() <= 0 {
		return nil, fmt.Errorf("argument %q should have array type", arg.Name)
	}
	var result []string

	for i := 0; i < arg.Value.Length(); i++ {
		el := arg.Value.Index(i)

		if el.Type() != js.TypeString {
			return nil, fmt.Errorf("argument %q array should contains only strings", arg.Name)
		}

		result = append(result, el.String())
	}

	return result, nil
}
