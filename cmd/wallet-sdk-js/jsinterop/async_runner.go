//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jsinterop

import (
	"fmt"
	"sync"
	"syscall/js"
)

type AsyncFunc func(this js.Value, args []js.Value) (any, error)

var (
	jsErr     js.Value = js.Global().Get("Error")
	jsPromise js.Value = js.Global().Get("Promise")
)

type AsyncRunner struct {
	serializeCallsMutex sync.Mutex
}

func (r *AsyncRunner) CreateAsyncFunc(innerFunc AsyncFunc) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(_ js.Value, promFn []js.Value) any {
			resolve, reject := promFn[0], promFn[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(jsErr.New(fmt.Sprint("panic:", r)))
					}
				}()

				// All Go function calls are invoked in goroutines to avoid blocking the JS side.
				// But to avoid race conditions inside Go code we serialize calls using mutex.
				r.serializeCallsMutex.Lock()

				res, err := innerFunc(this, args)

				r.serializeCallsMutex.Unlock()

				if err != nil {
					reject.Invoke(jsErr.New(err.Error()))
				} else {
					resolve.Invoke(res)
				}
			}()

			return nil
		})

		return jsPromise.New(handler)
	})
}
