package walleterror_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
)

func TestParse(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rawError := `{"code":"978", "category":"GENERAL_ERROR"}`

		err := walleterror.Parse(rawError)

		require.Equal(t, "978", err.Code)
		require.Equal(t, "GENERAL_ERROR", err.Category)
		require.Empty(t, err.Details)
	})
	t.Run("UnknownError", func(t *testing.T) {
		rawError := `{"code":978}`

		err := walleterror.Parse(rawError)

		require.Equal(t, "UKN2-000", err.Code)
		require.Equal(t, "OTHER_ERROR", err.Category)
		require.Equal(t, rawError, err.Details)
	})
}
