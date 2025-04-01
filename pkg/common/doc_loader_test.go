package common_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
)

func TestCreateJSONLDDocumentLoader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		loader, err := common.CreateJSONLDDocumentLoader(&http.Client{}, legacy.NewProvider())
		require.NoError(t, err)
		require.NotNil(t, loader)
	})

	t.Run("Fail context store", func(t *testing.T) {
		store := NewMockStoreProvider()
		store.FailNamespace = "ldcontexts"

		loader, err := common.CreateJSONLDDocumentLoader(&http.Client{}, store)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create JSON-LD context store")
		require.Nil(t, loader)
	})

	t.Run("Fail context store", func(t *testing.T) {
		store := NewMockStoreProvider()
		store.FailNamespace = "remoteproviders"

		loader, err := common.CreateJSONLDDocumentLoader(&http.Client{}, store)
		require.Error(t, err)
		require.Contains(t, err.Error(), "create remote provider store")
		require.Nil(t, loader)
	})
}
