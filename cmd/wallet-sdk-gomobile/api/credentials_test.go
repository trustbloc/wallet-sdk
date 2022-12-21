/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package api_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

var (
	//go:embed testdata/credential_university_degree.jsonld
	universityDegreeCredential []byte

	//go:embed testdata/credential_drivers_licence.jsonld
	driversLicenceCredential []byte
)

func TestVerifiableCredential(t *testing.T) {
	vcArray := api.NewVerifiableCredentialsArray()

	universityDegreeVC := api.NewVerifiableCredential(string(universityDegreeCredential))

	vcArray.Add(universityDegreeVC)

	require.Equal(t, 1, vcArray.Length())

	driversLicenceVC := api.NewVerifiableCredential(string(driversLicenceCredential))

	vcArray.Add(driversLicenceVC)

	require.Equal(t, 2, vcArray.Length())

	vc1 := vcArray.AtIndex(0)
	require.Equal(t, universityDegreeVC, vc1)

	vc2 := vcArray.AtIndex(1)
	require.Equal(t, driversLicenceVC, vc2)
}
