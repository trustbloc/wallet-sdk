/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package verifiable_test

import (
	"crypto/ed25519"
	"crypto/rand"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	afgojwt "github.com/trustbloc/vc-go/jwt"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

//go:embed testdata/credential_drivers_licence.jsonld
var driversLicenceCredential string

func TestVerifiableCredential(t *testing.T) {
	t.Run("Valid VCs", func(t *testing.T) {
		vcArray := verifiable.NewCredentialsArray()

		vcArray.AtIndex(0)

		opts := &verifiable.Opts{}
		opts.DisableProofCheck()

		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
		require.NoError(t, err)

		id := universityDegreeVC.ID()
		require.Equal(t, "http://example.edu/credentials/1872", id)

		name := universityDegreeVC.Name()
		require.Equal(t, "University Degree Credential", name)

		issuerID := universityDegreeVC.IssuerID()
		require.Equal(t, "did:example:76e12ec712ebc6f1c221ebfeb1f", issuerID)

		typesFromClaims := universityDegreeVC.ClaimTypes()
		require.Equal(t, 1, typesFromClaims.Length())
		require.Equal(t, "UniversityDegreeCredential", typesFromClaims.AtIndex(0))

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

		driversLicenceVC, err := verifiable.ParseCredential(driversLicenceCredential, opts)
		require.NoError(t, err)

		id = driversLicenceVC.ID()
		require.Equal(t, "https://eu.com/claims/DriversLicense", id)

		name = driversLicenceVC.Name()
		require.Empty(t, name)

		issuerID = driversLicenceVC.IssuerID()
		require.Equal(t, "did:foo:123", issuerID)

		typesFromClaims = driversLicenceVC.ClaimTypes()
		require.Equal(t, 1, typesFromClaims.Length())
		require.Equal(t, "DriversLicence", typesFromClaims.AtIndex(0))

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
		invalidVC := afgoverifiable.Credential{}

		vcWrapper := verifiable.NewCredential(&invalidVC)

		issuanceDate, err := vcWrapper.IssuanceDate()
		require.EqualError(t, err, "issuance date missing (invalid VC)")
		require.Equal(t, int64(-1), issuanceDate)
	})
}

func TestVerifiableCredential_NameIsNotAString(t *testing.T) {
	opts := &verifiable.Opts{}
	opts.DisableProofCheck()

	universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
	require.NoError(t, err)

	universityDegreeVC.VC.CustomFields["name"] = 0

	name := universityDegreeVC.Name()
	require.Empty(t, name)
}

func TestVerifiableCredential_ClaimTypes(t *testing.T) {
	t.Run("Claim types are in credential subject", func(t *testing.T) {
		opts := &verifiable.Opts{}
		opts.DisableProofCheck()

		t.Run("Types were an array of interfaces", func(t *testing.T) {
			universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
			require.NoError(t, err)

			types := universityDegreeVC.Types()
			require.Equal(t, 2, types.Length())
			require.Equal(t, "VerifiableCredential", types.AtIndex(0))
			require.Equal(t, "UniversityDegreeCredential", types.AtIndex(1))
		})
		t.Run("Type was a string", func(t *testing.T) {
			driversLicenceVC, err := verifiable.ParseCredential(driversLicenceCredential, opts)
			require.NoError(t, err)

			types := driversLicenceVC.Types()
			require.Equal(t, 1, types.Length())
			require.Equal(t, "VerifiableCredential", types.AtIndex(0))
		})
	})
	t.Run("Claim types are in selective disclosures", func(t *testing.T) {
		opts := []afgoverifiable.CredentialOpt{
			afgoverifiable.WithDisabledProofCheck(),
			afgoverifiable.WithCredDisableValidation(),
		}

		universityDegreeVC, err := afgoverifiable.ParseCredential([]byte(universityDegreeCredential), opts...)
		require.NoError(t, err)

		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		universityDegreeVCSDJWT, err := universityDegreeVC.MakeSDJWT(afgojwt.NewEd25519Signer(privKey),
			universityDegreeVC.Issuer.ID+"#keys-1")
		require.NoError(t, err)

		universityDegreeVCSD, err := afgoverifiable.ParseCredential([]byte(universityDegreeVCSDJWT), opts...)
		require.NoError(t, err)

		universityDegreeVCWithOnlyTypeSDJWT, err := universityDegreeVCSD.MarshalWithDisclosure(
			afgoverifiable.DiscloseGivenRequired([]string{"type"}))
		require.NoError(t, err)

		vcWithOnlyTypeSD, err := afgoverifiable.ParseCredential(
			[]byte(universityDegreeVCWithOnlyTypeSDJWT), opts...)
		require.NoError(t, err)

		wrappedUniversityDegreeVCWithOnlyTypeSD := verifiable.NewCredential(vcWithOnlyTypeSD)

		types := wrappedUniversityDegreeVCWithOnlyTypeSD.ClaimTypes()
		require.Equal(t, 1, types.Length())
		require.Equal(t, "UniversityDegreeCredential", types.AtIndex(0))
	})
	t.Run("No claim types", func(t *testing.T) {
		vc := verifiable.NewCredential(&afgoverifiable.Credential{})

		types := vc.ClaimTypes()
		require.Nil(t, types)
	})
}
