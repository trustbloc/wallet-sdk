/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package otel_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
)

func TestNewTrace(t *testing.T) {
	trace, err := otel.NewTrace()

	require.NotNil(t, trace)
	require.NoError(t, err)
	require.NotEmpty(t, trace.TraceID())
	require.NotNil(t, trace.TraceHeader())
}

func TestGenerateTrace(t *testing.T) {
	t.Run("Test header value", func(t *testing.T) {
		testTraceID := "1234567890abcdef"
		testSpanID := "12345678"

		trace, err := otel.GenerateTrace(func(b []byte) (int, error) {
			if len(b) == 16 {
				return copy(b, testTraceID), nil
			}

			if len(b) == 8 {
				return copy(b, testSpanID), nil
			}

			return 0, nil
		})

		require.NotNil(t, trace)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("00-%s-%s-01",
			hex.EncodeToString([]byte(testTraceID)),
			hex.EncodeToString([]byte(testSpanID))),
			trace.TraceHeader().Value)
	})

	t.Run("Trace id gen failed", func(t *testing.T) {
		trace, err := otel.GenerateTrace(func(b []byte) (int, error) {
			if len(b) == 16 {
				return 0, fmt.Errorf("trace id gen failed")
			}

			return 0, nil
		})

		require.Nil(t, trace)
		require.Error(t, err)
		require.Contains(t, err.Error(), "trace id gen failed")
	})

	t.Run("Span id gen failed", func(t *testing.T) {
		trace, err := otel.GenerateTrace(func(b []byte) (int, error) {
			if len(b) == 8 {
				return 0, fmt.Errorf("span id gen failed")
			}

			return 0, nil
		})

		require.Nil(t, trace)
		require.Error(t, err)
		require.Contains(t, err.Error(), "span id gen failed")
	})
}
