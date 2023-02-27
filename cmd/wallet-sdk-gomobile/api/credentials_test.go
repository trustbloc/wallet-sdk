/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package api_test

import (
	_ "embed"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api/vcparse"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/credential_university_degree.jsonld
	universityDegreeCredential string

	//go:embed testdata/credential_drivers_licence.jsonld
	driversLicenceCredential string
)

func TestVerifiableCredential(t *testing.T) {
	t.Run("Valid VCs", func(t *testing.T) {
		vcArray := api.NewVerifiableCredentialsArray()

		parseOpts := &vcparse.Opts{DisableProofCheck: true}

		universityDegreeVC, err := vcparse.Parse(universityDegreeCredential, parseOpts)
		require.NoError(t, err)

		id := universityDegreeVC.ID()
		require.Equal(t, "http://example.edu/credentials/1872", id)

		issuerID := universityDegreeVC.IssuerID()
		require.Equal(t, "did:example:76e12ec712ebc6f1c221ebfeb1f", issuerID)

		types := universityDegreeVC.Types()
		require.Equal(t, 2, types.Length())
		require.Equal(t, "VerifiableCredential", types.AtIndex(0))
		require.Equal(t, "UniversityDegreeCredential", types.AtIndex(1))

		issuanceDate, err := universityDegreeVC.IssuanceDate()
		require.NoError(t, err)
		require.Equal(t, int64(1262373804), issuanceDate)

		hasExpirationDate := universityDegreeVC.HasExpirationDate()
		require.True(t, hasExpirationDate)

		expirationDate, err := universityDegreeVC.ExpirationDate()
		require.NoError(t, err)
		require.Equal(t, int64(1577906604), expirationDate)

		vcArray.Add(universityDegreeVC)

		require.Equal(t, 1, vcArray.Length())

		driversLicenceVC, err := vcparse.Parse(driversLicenceCredential, parseOpts)
		require.NoError(t, err)

		id = driversLicenceVC.ID()
		require.Equal(t, "https://eu.com/claims/DriversLicense", id)

		issuerID = driversLicenceVC.IssuerID()
		require.Equal(t, "did:foo:123", issuerID)

		types = driversLicenceVC.Types()
		require.Equal(t, 1, types.Length())
		require.Equal(t, "VerifiableCredential", types.AtIndex(0))

		issuanceDate, err = driversLicenceVC.IssuanceDate()
		require.NoError(t, err)
		require.Equal(t, int64(1262375604), issuanceDate)

		hasExpirationDate = driversLicenceVC.HasExpirationDate()
		require.False(t, hasExpirationDate)

		expirationDate, err = driversLicenceVC.ExpirationDate()
		require.EqualError(t, err, "VC has no expiration date")
		require.Equal(t, int64(-1), expirationDate)

		vcArray.Add(driversLicenceVC)

		require.Equal(t, 2, vcArray.Length())

		vc1 := vcArray.AtIndex(0)
		require.Equal(t, universityDegreeVC, vc1)

		vc2 := vcArray.AtIndex(1)
		require.Equal(t, driversLicenceVC, vc2)

		serializedVC, err := universityDegreeVC.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedVC)
	})
	t.Run("Invalid VC", func(t *testing.T) {
		invalidVC := verifiable.Credential{}

		vcWrapper := api.NewVerifiableCredential(&invalidVC)

		issuanceDate, err := vcWrapper.IssuanceDate()
		require.EqualError(t, err, "issuance date missing (invalid VC)")
		require.Equal(t, int64(-1), issuanceDate)
	})
}
