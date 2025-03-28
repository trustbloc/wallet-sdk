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
		require.Equal(t, "#12107c", issuer1.BackgroundColor())
		require.Equal(t, "#FFFFFF", issuer1.TextColor())
		require.Equal(t, "https://exampleuniversity.com/public/logo.png", issuer1.Logo().URL())
		require.Equal(t, "a square logo of a university", issuer1.Logo().AltText())

		_, err = issuer1.Serialize()
		require.NoError(t, err)

		issuer2 := resolvedData.LocalizedIssuerAtIndex(1)
		require.Equal(t, "サンプル大学", issuer2.Name())
		require.Equal(t, "jp-JA", issuer2.Locale())
		require.Equal(t, "", issuer2.URL())
		require.Nil(t, issuer2.Logo())

		require.Nil(t, resolvedData.LocalizedIssuerAtIndex(-1))

		credential1 := resolvedData.CredentialAtIndex(0)
		require.Equal(t, 1, credential1.LocalizedOverviewsLength())
		require.Equal(t, 6, credential1.SubjectsLength())

		require.Nil(t, resolvedData.CredentialAtIndex(-1))

		overview1 := credential1.LocalizedOverviewAtIndex(0)
		require.Equal(t, "University Credential", overview1.Name())
		require.Equal(t, "en-US", overview1.Locale())
		require.Equal(t, "#12107c", overview1.BackgroundColor())
		require.Equal(t, "#FFFFFF", overview1.TextColor())
		require.Equal(t, "https://exampleuniversity.com/public/logo.png", overview1.Logo().URL())
		require.Equal(t, "a square logo of a university", overview1.Logo().AltText())

		require.Nil(t, credential1.LocalizedOverviewAtIndex(-1))

		subject1 := credential1.SubjectAtIndex(0)
		require.Equal(t, "4.0", subject1.Value())
		require.Equal(t, `^\d+(\.\d+)?$`, subject1.Pattern())

		_, err = subject1.Order()
		require.Error(t, err)

		require.Nil(t, credential1.SubjectAtIndex(-1))

		subject2 := credential1.SubjectAtIndex(1)
		require.Equal(t, "sensitive_id", subject2.RawID())
		require.Equal(t, "string", subject2.ValueType())
		require.Equal(t, "123456789", subject2.RawValue())
		require.Equal(t, "•••••6789", subject2.Value())
		require.True(t, subject2.IsMasked())
		require.Empty(t, subject2.Attachment())

		require.Equal(t, 1, subject2.LocalizedLabelsLength())
		subject2LocalizedLabels := subject2.LocalizedLabelAtIndex(0)
		require.Equal(t, "Sensitive ID", subject2LocalizedLabels.Name())
		require.Equal(t, "en-US", subject2LocalizedLabels.Locale())

		require.Nil(t, subject2.LocalizedLabelAtIndex(-1))

		require.True(t, subject2.HasOrder())
		subject2Order, err := subject2.Order()
		require.NoError(t, err)
		require.Equal(t, 5, subject2Order)
	})
	t.Run("Serialize", func(t *testing.T) {
		resolvedData, err := display.ParseResolvedData(universityDegreeResolvedData)
		require.NoError(t, err)

		serialized, err := resolvedData.Serialize()
		require.NoError(t, err)

		require.JSONEq(t, universityDegreeResolvedData, serialized)
	})
}
