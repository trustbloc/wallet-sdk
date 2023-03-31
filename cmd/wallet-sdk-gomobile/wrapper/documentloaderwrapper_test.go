/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
)

func TestDocumentLoaderWrapper_LoadDocument(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		documentLoaderWrapper := wrapper.DocumentLoaderWrapper{
			DocumentLoader: &documentLoaderMock{
				LoadResult: &api.LDDocument{
					DocumentURL: "testURL",
					Document:    "{}",
					ContextURL:  "testContext",
				},
			},
		}

		resultDoc, err := documentLoaderWrapper.LoadDocument("testUrl")
		require.NoError(t, err)
		require.NotNil(t, resultDoc.Document)
		require.Equal(t, "testURL", resultDoc.DocumentURL)
		require.Equal(t, "testContext", resultDoc.ContextURL)
	})

	t.Run("Load failed", func(t *testing.T) {
		documentLoaderWrapper := wrapper.DocumentLoaderWrapper{
			DocumentLoader: &documentLoaderMock{
				LoadErr: errors.New("load failed"),
			},
		}

		_, err := documentLoaderWrapper.LoadDocument("testUrl")
		require.Error(t, err)
		require.Contains(t, err.Error(), "load failed")
	})

	t.Run("Doc parse failed", func(t *testing.T) {
		documentLoaderWrapper := wrapper.DocumentLoaderWrapper{
			DocumentLoader: &documentLoaderMock{
				LoadResult: &api.LDDocument{
					DocumentURL: "testURL",
					Document:    "",
					ContextURL:  "testContext",
				},
			},
		}

		_, err := documentLoaderWrapper.LoadDocument("testUrl")
		require.Error(t, err)
		require.Contains(t, err.Error(), "fail to unmarshal ld document bytes")
	})
}

type documentLoaderMock struct {
	LoadResult *api.LDDocument
	LoadErr    error
}

func (d *documentLoaderMock) LoadDocument(string) (*api.LDDocument, error) {
	return d.LoadResult, d.LoadErr
}
