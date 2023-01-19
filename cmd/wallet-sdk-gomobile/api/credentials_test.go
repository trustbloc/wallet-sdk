/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package api_test

import (
	_ "embed"
	"testing"

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
	vcArray := api.NewVerifiableCredentialsArray()

	parseOpts := &vcparse.Opts{DisableProofCheck: true}

	universityDegreeVC, err := vcparse.Parse(universityDegreeCredential, parseOpts)
	require.NoError(t, err)

	id := universityDegreeVC.IssuerID()
	require.Equal(t, "did:example:76e12ec712ebc6f1c221ebfeb1f", id)

	vcArray.Add(universityDegreeVC)

	require.Equal(t, 1, vcArray.Length())

	driversLicenceVC, err := vcparse.Parse(driversLicenceCredential, parseOpts)
	require.NoError(t, err)

	id = driversLicenceVC.IssuerID()
	require.Equal(t, "did:foo:123", id)

	vcArray.Add(driversLicenceVC)

	require.Equal(t, 2, vcArray.Length())

	vc1 := vcArray.AtIndex(0)
	require.Equal(t, universityDegreeVC, vc1)

	vc2 := vcArray.AtIndex(1)
	require.Equal(t, driversLicenceVC, vc2)

	serializedVC, err := universityDegreeVC.Serialize()
	require.NoError(t, err)
	require.NotEmpty(t, serializedVC)
}
