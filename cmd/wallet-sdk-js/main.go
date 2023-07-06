//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"syscall/js"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop"
)

const (
	agentJSObjectName  = "__agentInteropObject"
	stopAssemblyFnName = "stopAssembly"
)

var done = make(chan struct{}, 0)

func main() {

	exports := jsinterop.ExportAgentFunctions()
	exports[stopAssemblyFnName] = js.FuncOf(stopAssembly)

	js.Global().Set(agentJSObjectName, exports)

	<-done
}

func stopAssembly(_ js.Value, _ []js.Value) interface{} {
	done <- struct{}{}

	return nil
}
