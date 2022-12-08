/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
)

var (
	//go:embed test_data/presentation_definition.json
	presentationDefinition []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVCJWT string

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC string
)

func TestInstance_Query(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		query := credential.NewInquirer(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		presentation, err := query.Query(presentationDefinition,
			createCredJSONArray(t, []string{universityDegreeVCJWT, permanentResidentCardVC}),
		)
		require.NoError(t, err)
		require.NotNil(t, presentation)
	})

	t.Run("No matched credential", func(t *testing.T) {
		query := credential.NewInquirer(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition,
			createCredJSONArray(t, []string{permanentResidentCardVC}),
		)
		require.Contains(t, err.Error(), "credentials do not satisfy requirements")
	})

	t.Run("PD parse failed", func(t *testing.T) {
		query := credential.NewInquirer(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(nil,
			createCredJSONArray(t, []string{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "unmarshal of presentation definition failed:")
	})

	t.Run("PD validation failed", func(t *testing.T) {
		query := credential.NewInquirer(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query([]byte("{}"),
			createCredJSONArray(t, []string{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "validation of presentation definition failed:")
	})

	t.Run("VC parse failed", func(t *testing.T) {
		query := credential.NewInquirer(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition,
			createCredJSONArray(t, []string{"{}"}),
		)

		require.Contains(t, err.Error(), "verifiable credential parse failed")
	})

	t.Run("Nil credentials and nil reader", func(t *testing.T) {
		query := credential.NewInquirer(&documentLoaderReverseWrapper{
			DocumentLoader: testutil.DocumentLoader(t),
		})

		_, err := query.Query(presentationDefinition, credential.NewCredentialsOptFromReader(nil))

		require.Contains(t, err.Error(), "either credential reader or vc array should be set")
	})
}

func createCredJSONArray(t *testing.T, creds []string) *credential.CredentialsOpt {
	t.Helper()

	credsArray := api.NewVerifiableCredentialsArray()
	for _, credContent := range creds {
		credsArray.Add(api.NewVerifiableCredential([]byte(credContent)))
	}

	return credential.NewCredentialsOpt(credsArray)
}

type documentLoaderReverseWrapper struct {
	DocumentLoader ld.DocumentLoader
}

func (l *documentLoaderReverseWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	wrappedDoc := &api.LDDocument{
		DocumentURL: doc.DocumentURL,
		ContextURL:  doc.ContextURL,
	}

	wrappedDoc.Document, err = json.Marshal(doc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	return wrappedDoc, nil
}
