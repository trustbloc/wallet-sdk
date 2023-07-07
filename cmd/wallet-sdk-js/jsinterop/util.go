//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jsinterop

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

type namedArgument struct {
	name string
	val  js.Value
}

func newNamedArg(name string, val js.Value) namedArgument {
	return namedArgument{
		name: name,
		val:  val,
	}
}

// Function with named arguments in JS will look like this someFunc({param1: "value"})
func getNamedArgument(args []js.Value, name string) (namedArgument, error) {
	if len(args) > 1 {
		return newNamedArg(name, js.Null()), walleterror.NewValidationError(
			Module,
			InvalidArgumentsCode,
			InvalidArgumentsError,
			errors.New("function with named parameters should receive only one or zero arguments"))
	}

	if len(args) == 1 {
		val := args[0].Get(name)

		if !val.IsNull() && !val.IsUndefined() {
			return newNamedArg(name, val), nil
		}
	}

	return newNamedArg(name, js.Null()),
		walleterror.NewValidationError(
			Module,
			InvalidArgumentsCode,
			InvalidArgumentsError,
			fmt.Errorf("named parameter %s is required", name))
}

func getOptionalNamedArgument(args []js.Value, name string) (namedArgument, error) {
	if len(args) > 1 {
		return newNamedArg(name, js.Null()), walleterror.NewValidationError(
			Module,
			InvalidArgumentsCode,
			InvalidArgumentsError,
			errors.New("function with named parameters should receive only one or zero arguments"))
	}

	if len(args) == 1 {
		val := args[0].Get(name)

		if !val.IsNull() && !val.IsUndefined() {
			return newNamedArg(name, val), nil
		}
	}

	return newNamedArg(name, js.Null()), nil
}

func ensureString(arg namedArgument, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if arg.val.Type() != js.TypeString {
		return "", fmt.Errorf("argument %q should have type string", arg.name)
	}

	return arg.val.String(), nil
}
