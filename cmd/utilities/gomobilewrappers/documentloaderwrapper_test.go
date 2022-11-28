/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gomobilewrappers_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/utilities/gomobilewrappers"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

func TestDocumentLoaderWrapper_LoadDocument(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wrapper := gomobilewrappers.DocumentLoaderWrapper{
			DocumentLoader: &documentLoaderMock{
				LoadResult: &api.LDDocument{
					DocumentURL: "testURL",
					Document:    []byte("{}"),
					ContextURL:  "testContext",
				},
			},
		}

		resultDoc, err := wrapper.LoadDocument("testUrl")
		require.NoError(t, err)
		require.NotNil(t, resultDoc.Document)
		require.Equal(t, "testURL", resultDoc.DocumentURL)
		require.Equal(t, "testContext", resultDoc.ContextURL)
	})

	t.Run("Load failed", func(t *testing.T) {
		wrapper := gomobilewrappers.DocumentLoaderWrapper{
			DocumentLoader: &documentLoaderMock{
				LoadErr: errors.New("load failed"),
			},
		}

		_, err := wrapper.LoadDocument("testUrl")
		require.Error(t, err)
		require.Contains(t, err.Error(), "load failed")
	})

	t.Run("DOc parse failed", func(t *testing.T) {
		wrapper := gomobilewrappers.DocumentLoaderWrapper{
			DocumentLoader: &documentLoaderMock{
				LoadResult: &api.LDDocument{
					DocumentURL: "testURL",
					Document:    nil,
					ContextURL:  "testContext",
				},
			},
		}

		_, err := wrapper.LoadDocument("testUrl")
		require.Error(t, err)
		require.Contains(t, err.Error(), "fail to unmarshal ld document bytes")
	})
}

type documentLoaderMock struct {
	LoadResult *api.LDDocument
	LoadErr    error
}

func (d *documentLoaderMock) LoadDocument(u string) (*api.LDDocument, error) {
	return d.LoadResult, d.LoadErr
}
