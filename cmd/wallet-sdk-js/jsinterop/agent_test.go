//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jsinterop_test

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop"
)

func TestInitAgent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		opts := map[string]any{
			"didResolverURI": "",
		}

		_, err := jsinterop.initAgent(js.Null(), []js.Value{js.ValueOf(opts)})
		require.NoError(t, err)
	})
}
