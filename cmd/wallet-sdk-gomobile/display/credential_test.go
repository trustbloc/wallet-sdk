package display_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
)

func TestParseResolvedData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resolvedData, err := display.ParseResolvedData(universityDegreeResolvedData)
		require.NoError(t, err)

		require.Equal(t, 2, resolvedData.LocalizedIssuersLength())
		require.Equal(t, 1, resolvedData.CredentialsLength())

		issuer1 := resolvedData.LocalizedIssuerAtIndex(0)
		require.Equal(t, "Example University", issuer1.Name())
		require.Equal(t, "en-US", issuer1.Locale())
		require.Equal(t, "https://server.example.com", issuer1.URL())

		issuer2 := resolvedData.LocalizedIssuerAtIndex(1)
		require.Equal(t, "サンプル大学", issuer2.Name())
		require.Equal(t, "jp-JA", issuer2.Locale())
		require.Equal(t, "", issuer2.URL())

		credential1 := resolvedData.CredentialAtIndex(0)
		require.Equal(t, 1, credential1.LocalizedOverviewsLength())
		require.Equal(t, 6, credential1.SubjectsLength())

		overview1 := credential1.LocalizedOverviewAtIndex(0)
		require.Equal(t, "University Credential", overview1.Name())
		require.Equal(t, "en-US", overview1.Locale())
		require.Equal(t, "#12107c", overview1.BackgroundColor())
		require.Equal(t, "#FFFFFF", overview1.TextColor())
		require.Equal(t, "https://exampleuniversity.com/public/logo.png", overview1.Logo().URL())
		require.Equal(t, "a square logo of a university", overview1.Logo().AltText())
	})
	t.Run("Serialize", func(t *testing.T) {
		resolvedData, err := display.ParseResolvedData(universityDegreeResolvedData)
		require.NoError(t, err)

		serialized, err := resolvedData.Serialize()
		require.NoError(t, err)

		require.JSONEq(t, universityDegreeResolvedData, serialized)
	})
}
