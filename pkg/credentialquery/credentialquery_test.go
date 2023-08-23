/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialquery_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/vc-go/did"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
)

var (
	//go:embed test_data/multi_inputs_pd.json
	multiInputPD []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVC []byte

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC []byte

	//go:embed test_data/drivers_license.jwt
	driverLicenseVC []byte

	//go:embed test_data/verified_employee.jwt
	verifiedEmployeeVC []byte
)

func TestInstance_GetSubmissionRequirements(t *testing.T) {
	docLoader := testutil.DocumentLoader(t)
	pdQuery := &presexch.PresentationDefinition{}
	err := json.Unmarshal(multiInputPD, pdQuery)
	require.NoError(t, err)

	contents := [][]byte{
		universityDegreeVC,
		permanentResidentCardVC,
		driverLicenseVC,
		verifiedEmployeeVC,
	}

	var credentials []*verifiable.Credential

	for _, credContent := range contents {
		cred, credErr := verifiable.ParseCredential(credContent, verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(docLoader))
		require.NoError(t, credErr)

		credentials = append(credentials, cred)
	}

	t.Run("Success", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		requirements, err := instance.GetSubmissionRequirements(pdQuery, credentialquery.WithCredentialsArray(
			credentials,
		))

		require.NoError(t, err)
		require.Len(t, requirements, 1)

		require.Len(t, requirements[0].Descriptors, 3)
	})

	t.Run("Enable selective disclosure", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		requirements, err := instance.GetSubmissionRequirements(pdQuery, credentialquery.WithCredentialsArray(
			credentials,
		), credentialquery.WithSelectiveDisclosure(&didResolverMock{}))

		require.NoError(t, err)
		require.Len(t, requirements, 1)

		require.Len(t, requirements[0].Descriptors, 3)
	})

	t.Run("Reader error", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		_, err := instance.GetSubmissionRequirements(pdQuery, credentialquery.WithCredentialReader(
			&readerMock{
				err: errors.New("get all error"),
			},
		))

		require.Error(t, err, "credential reader failed: get all error")
	})

	t.Run("Credentials not provided", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		_, err := instance.GetSubmissionRequirements(pdQuery)

		testutil.RequireErrorContains(t, err, "CREDENTIAL_READER_NOT_SET")
	})

	t.Run("Checks schema", func(t *testing.T) {
		incorrectPD := &presexch.PresentationDefinition{ID: uuid.New().String()}

		instance := credentialquery.NewInstance(docLoader)
		requirements, err := instance.GetSubmissionRequirements(incorrectPD, credentialquery.WithCredentialsArray(
			credentials,
		))

		testutil.RequireErrorContains(t, err, "FAIL_TO_GET_MATCH_REQUIREMENTS_RESULTS")
		require.Nil(t, requirements)
	})
}

type readerMock struct {
	credentials []*verifiable.Credential
	err         error
}

func (r *readerMock) Get(string) (*verifiable.Credential, error) {
	return nil, r.err
}

func (r *readerMock) GetAll() ([]*verifiable.Credential, error) {
	return r.credentials, r.err
}

type didResolverMock struct {
	ResolveValue *did.DocResolution
	ResolveErr   error
}

func (d *didResolverMock) Resolve(string) (*did.DocResolution, error) {
	return d.ResolveValue, d.ResolveErr
}
