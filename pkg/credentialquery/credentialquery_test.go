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

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
)

var (
	//go:embed test_data/presentation_definition.json
	presentationDefinition []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVC []byte

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC []byte
)

func TestInstance_Query(t *testing.T) {
	docLoader := testutil.DocumentLoader(t)
	pdQuery := &presexch.PresentationDefinition{}
	err := json.Unmarshal(presentationDefinition, pdQuery)
	require.NoError(t, err)

	contents := [][]byte{
		universityDegreeVC,
		permanentResidentCardVC,
	}

	var credentials []*verifiable.Credential

	for _, credContent := range contents {
		cred, credErr := verifiable.ParseCredential(credContent, verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(docLoader))
		require.NoError(t, credErr)

		credentials = append(credentials, cred)
	}

	t.Run("Success with vc array", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		presentation, err := instance.Query(pdQuery, credentialquery.WithCredentialsArray(
			credentials,
		))

		require.NoError(t, err)
		require.NotNil(t, presentation)

		require.Len(t, presentation.Credentials(), 1)
	})

	t.Run("Success with reader", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		presentation, err := instance.Query(pdQuery, credentialquery.WithCredentialReader(
			&readerMock{
				credentials: credentials,
			},
		))

		require.NoError(t, err)
		require.NotNil(t, presentation)

		require.Len(t, presentation.Credentials(), 1)
	})

	t.Run("Reader error", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		_, err := instance.Query(pdQuery, credentialquery.WithCredentialReader(
			&readerMock{
				err: errors.New("get all error"),
			},
		))

		require.Error(t, err, "credential reader failed: get all error")
	})

	t.Run("Credential reader not set.", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		_, err := instance.Query(pdQuery)

		testutil.RequireErrorContains(t, err, "CREDENTIAL_READER_NOT_SET")
	})

	t.Run("No Credentials", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)
		_, err := instance.Query(pdQuery, credentialquery.WithCredentialsArray(
			[]*verifiable.Credential{{}},
		))

		testutil.RequireErrorContains(t, err, "NO_CREDENTIAL_SATISFY_REQUIREMENTS")
	})

	t.Run("Create vp failed", func(t *testing.T) {
		instance := credentialquery.NewInstance(docLoader)

		_, err := instance.Query(&presexch.PresentationDefinition{}, credentialquery.WithCredentialsArray(
			credentials,
		))

		testutil.RequireErrorContains(t, err, "CREATE_VP_FAILED")
	})
}

type readerMock struct {
	credentials []*verifiable.Credential
	err         error
}

func (r *readerMock) Get(id string) (*verifiable.Credential, error) {
	return nil, r.err
}

func (r *readerMock) GetAll() ([]*verifiable.Credential, error) {
	return r.credentials, r.err
}
