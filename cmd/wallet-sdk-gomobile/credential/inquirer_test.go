/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

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

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api/vcparse"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
)

var (
	//go:embed test_data/presentation_definition.json
	presentationDefinition []byte

	//go:embed test_data/multi_inputs_pd.json
	multiInputPD []byte

	//go:embed test_data/nested_submission_requirements_pd.json
	nestedRequirementsPD []byte

	//go:embed test_data/university_degree.jwt
	universityDegreeVCJWT []byte

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC []byte

	//go:embed test_data/drivers_license.jwt
	driverLicenseVC []byte

	//go:embed test_data/verified_employee.jwt
	verifiedEmployeeVC []byte
)

func TestNewInquirer(t *testing.T) {
	t.Run("Using the default network-based document loader", func(t *testing.T) {
		inquirer := credential.NewInquirer(nil)
		require.NotNil(t, inquirer)
	})
}

func TestInstance_Query(t *testing.T) {
	opts := credential.NewInquirerOpts()

	documentLoader := &documentLoaderReverseWrapper{
		DocumentLoader: testutil.DocumentLoader(t),
	}

	opts.SetDocumentLoader(documentLoader)

	t.Run("Success", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		presentation, err := query.Query(presentationDefinition,
			createCredJSONArray(t, [][]byte{universityDegreeVCJWT, permanentResidentCardVC}),
		)
		require.NoError(t, err)
		require.NotNil(t, presentation)

		content, err := presentation.Content()
		require.NoError(t, err)
		require.NotNil(t, content)

		credentials, err := presentation.Credentials()
		require.NoError(t, err)
		require.Equal(t, 1, credentials.Length())
	})

	t.Run("No matched credential", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		_, err := query.Query(presentationDefinition,
			createCredJSONArray(t, [][]byte{permanentResidentCardVC}),
		)
		require.Contains(t, err.Error(), "credentials do not satisfy requirements")
	})

	t.Run("PD parse failed", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		_, err := query.Query(nil,
			createCredJSONArray(t, [][]byte{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "unmarshal of presentation definition failed:")
	})

	t.Run("PD validation failed", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		_, err := query.Query([]byte("{}"),
			createCredJSONArray(t, [][]byte{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "validation of presentation definition failed:")
	})

	t.Run("Nil credentials and nil reader", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		_, err := query.Query(presentationDefinition, credential.NewCredentialsArgFromReader(nil))

		require.Contains(t, err.Error(), "either credential reader or vc array must be set")
	})
}

func TestInstance_GetSubmissionRequirements(t *testing.T) {
	contents := [][]byte{
		universityDegreeVCJWT,
		permanentResidentCardVC,
		driverLicenseVC,
		verifiedEmployeeVC,
	}

	opts := credential.NewInquirerOpts()

	documentLoader := &documentLoaderReverseWrapper{
		DocumentLoader: testutil.DocumentLoader(t),
	}

	opts.SetDocumentLoader(documentLoader)

	t.Run("Success", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		requirements, err := query.GetSubmissionRequirements(multiInputPD, createCredJSONArray(t, contents))

		require.NoError(t, err)
		require.Equal(t, requirements.Len(), 1)
		req1 := requirements.AtIndex(0)
		require.Equal(t, req1.DescriptorLen(), 3)
		require.Equal(t, req1.Name(), "Information")
		require.Equal(t, req1.Purpose(), "test purpose")
		require.Equal(t, req1.Rule(), "pick")

		require.Equal(t, req1.Count(), 1)
		require.Equal(t, req1.Min(), 0)
		require.Equal(t, req1.Max(), 0)
		require.Equal(t, req1.NestedRequirementLength(), 0)

		desc1 := req1.DescriptorAtIndex(0)

		require.Equal(t, desc1.ID, "VerifiedEmployee")
		require.Equal(t, desc1.Name, "Verified Employee")
		require.Equal(t, desc1.Purpose, "test purpose")
		require.Equal(t, desc1.MatchedVCs.Length(), 1)
	})

	t.Run("Success nested requirements", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		requirements, err := query.GetSubmissionRequirements(nestedRequirementsPD, createCredJSONArray(t, contents))

		require.NoError(t, err)
		require.Equal(t, requirements.Len(), 1)
		req1 := requirements.AtIndex(0)
		require.Equal(t, req1.DescriptorLen(), 0)
		require.Equal(t, req1.Name(), "Nested requirements")
		require.Equal(t, req1.Rule(), "all")

		require.Equal(t, req1.Count(), 2)
		require.Equal(t, req1.Min(), 0)
		require.Equal(t, req1.Max(), 0)
		require.Equal(t, req1.NestedRequirementLength(), 2)

		nestedReq1 := req1.NestedRequirementAtIndex(0)

		require.Equal(t, nestedReq1.DescriptorLen(), 2)

		desc1 := nestedReq1.DescriptorAtIndex(0)

		require.Equal(t, desc1.ID, "VerifiedEmployee")
		require.Equal(t, desc1.Name, "Verified Employee")
		require.Equal(t, desc1.MatchedVCs.Length(), 1)
	})

	t.Run("PD parse failed", func(t *testing.T) {
		query := credential.NewInquirer(opts)

		_, err := query.GetSubmissionRequirements(nil,
			createCredJSONArray(t, [][]byte{universityDegreeVCJWT, permanentResidentCardVC}),
		)

		require.Contains(t, err.Error(), "unmarshal of presentation definition failed:")
	})
}

func createCredJSONArray(t *testing.T, creds [][]byte) *credential.CredentialsArg {
	t.Helper()

	credsArray := api.NewVerifiableCredentialsArray()

	for _, credContent := range creds {
		opts := vcparse.NewOpts()
		opts.DisableProofCheck()

		vc, err := vcparse.Parse(string(credContent), opts)
		require.NoError(t, err)

		credsArray.Add(vc)
	}

	return credential.NewCredentialsArgFromVCArray(credsArray)
}

type documentLoaderReverseWrapper struct {
	DocumentLoader ld.DocumentLoader
}

func (l *documentLoaderReverseWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	documentBytes, err := json.Marshal(doc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	wrappedDoc := &api.LDDocument{
		DocumentURL: doc.DocumentURL,
		Document:    string(documentBytes),
		ContextURL:  doc.ContextURL,
	}

	return wrappedDoc, nil
}
