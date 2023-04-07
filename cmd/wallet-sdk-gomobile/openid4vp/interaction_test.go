/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	"crypto/ed25519"
	"crypto/rand"
	_ "embed" //nolint:gci // required for go:embed
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	afgoverifiable "github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/models"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/credentials.jsonld
	credentialsJSONLD []byte
)

type mockVerifierServerHandler struct {
	t              *testing.T
	headersToCheck *api.Headers
}

// Simply checks the headers and return an arbitrary invalid response.
func (m *mockVerifierServerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, headerToCheck := range m.headersToCheck.GetAll() {
		// Note: for these tests, we're assuming that there aren't multiple values under a single name/key.
		value := request.Header.Get(headerToCheck.Name)
		require.Equal(m.t, headerToCheck.Value, value)
	}

	_, err := writer.Write([]byte("invalid"))
	require.NoError(m.t, err)
}

func TestOpenID4VP_GetQuery(t *testing.T) {
	t.Run("NewInteraction success", func(t *testing.T) {
		t.Run("Without any optional args", func(t *testing.T) {
			requiredArgs := NewArgs(
				requestObjectJWT,
				&mockCrypto{},
				&mocksDIDResolver{},
			)

			instance, err := NewInteraction(requiredArgs, nil)
			require.NoError(t, err)
			require.NotNil(t, instance)
			require.NotNil(t, instance.crypto)
			require.NotNil(t, instance.goAPIOpenID4VP)
			require.NotEmpty(t, instance.OTelTraceID())
		})
		t.Run("With optional args", func(t *testing.T) {
			requiredArgs := NewArgs(
				requestObjectJWT,
				&mockCrypto{},
				&mocksDIDResolver{},
			)

			// Note: in-depth testing of opts functionality is done in the integration tests.
			opts := NewOpts()
			opts.SetDocumentLoader(nil)
			opts.SetActivityLogger(nil)
			opts.SetMetricsLogger(nil)
			opts.DisableHTTPClientTLSVerify()
			opts.DisableOpenTelemetry()

			instance, err := NewInteraction(requiredArgs, opts)
			require.NoError(t, err)
			require.NotNil(t, instance)
			require.Empty(t, instance.OTelTraceID())
		})
	})

	t.Run("GetQuery success", func(t *testing.T) {
		t.Run("Without additional headers", func(t *testing.T) {
			instance := &Interaction{
				goAPIOpenID4VP: &mocGoAPIInteraction{
					GetQueryResult: &presexch.PresentationDefinition{},
				},
			}

			query, err := instance.GetQuery()
			require.NoError(t, err)
			require.NotNil(t, query)
		})
	})

	t.Run("GetQuery failed", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mocGoAPIInteraction{
				GetQueryError: errors.New("get query failed"),
			},
		}

		query, err := instance.GetQuery()
		require.Contains(t, err.Error(), "get query failed")
		require.Nil(t, query)
	})

	t.Run("With additional headers, and the server receives them", func(t *testing.T) {
		additionalHeaders := api.NewHeaders()

		additionalHeaders.Add(api.NewHeader("header-name-1", "header-value-1"))
		additionalHeaders.Add(api.NewHeader("header-name-2", "header-value-2"))

		opts := NewOpts()
		opts.AddHeaders(additionalHeaders)

		mockServer := &mockVerifierServerHandler{t: t, headersToCheck: additionalHeaders}
		testServer := httptest.NewServer(mockServer)

		defer testServer.Close()

		requiredArgs := NewArgs(
			"openid-vc://?request_uri="+testServer.URL,
			&mockCrypto{},
			&mocksDIDResolver{},
		)

		instance, err := NewInteraction(requiredArgs, opts)
		require.NoError(t, err)

		// The purpose of this test is to make sure the mock server receives the additional headers
		// as set above. It doesn't return a valid response, hence why GetQuery still fails.
		// If the server doesn't receive the headers as expected, the server itself will fail the
		// test (see the mockVerifierServerHandler.ServeHTTP method).
		// Any other error being returned by GetQuery is unexpected.
		query, err := instance.GetQuery()
		require.Error(t, err)
		require.Contains(t, err.Error(), `{"code":"OVP1-0001","category":`+
			`"VERIFY_AUTHORIZATION_REQUEST_FAILED","details":"verify authorization request: `+
			`parse JWT: JWT of compacted JWS form is supported only"`)
		require.Nil(t, query)
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	mockKey, _, e := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, e)

	credentials := verifiable.NewCredentialsArray()

	credentialData := []json.RawMessage{}

	e = json.Unmarshal(credentialsJSONLD, &credentialData)
	require.NoError(t, e)

	for _, credBytes := range credentialData {
		cred, err := afgoverifiable.ParseCredential(credBytes,
			afgoverifiable.WithDisabledProofCheck(), afgoverifiable.WithCredDisableValidation())
		require.NoError(t, err)

		credentials.Add(verifiable.NewCredential(cred))
	}

	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				PresentCredentialErr: nil,
			},
			didResolver: &mocksDIDResolver{ResolveDocBytes: mockResolution(t, &api.VerificationMethod{
				ID:   "did:example:12345#testId",
				Type: "Ed25519VerificationKey2018",
				Key:  models.VerificationKey{Raw: mockKey},
			})},
		}

		err := instance.PresentCredential(credentials)
		require.NoError(t, err)
	})

	t.Run("Present credentials failed", func(t *testing.T) {
		instance := &Interaction{
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mocGoAPIInteraction{
				PresentCredentialErr: errors.New("present credentials failed"),
			},
			didResolver: &mocksDIDResolver{ResolveDocBytes: mockResolution(t, &api.VerificationMethod{
				ID:   "did:example:12345#testId",
				Type: "Ed25519VerificationKey2018",
				Key:  models.VerificationKey{Raw: mockKey},
			})},
		}

		err := instance.PresentCredential(credentials)
		require.Contains(t, err.Error(), "present credentials failed")
	})
}

func TestInteraction_VerifierDisplayData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mocGoAPIInteraction{
				VerifierDisplayDataRes: &openid4vp.VerifierDisplayData{
					Name: "testName",
				},
			},
		}

		data, err := instance.VerifierDisplayData()
		require.NoError(t, err)
		require.Equal(t, "testName", data.Name)
	})

	t.Run("Error", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mocGoAPIInteraction{
				VerifierDisplayDataError: errors.New("testErr"),
			},
		}

		_, err := instance.VerifierDisplayData()
		require.Error(t, err)
		require.Contains(t, err.Error(), "testErr")
	})
}

type documentLoaderWrapper struct {
	goAPIDocumentLoader ld.DocumentLoader
}

func (dl *documentLoaderWrapper) LoadDocument(u string) (*api.LDDocument, error) {
	ldDoc, err := dl.goAPIDocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	docBytes, err := json.Marshal(ldDoc.Document)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ld document: %w", err)
	}

	return &api.LDDocument{
		DocumentURL: ldDoc.DocumentURL,
		Document:    string(docBytes),
		ContextURL:  ldDoc.ContextURL,
	}, nil
}

type mockCrypto struct {
	SignResult []byte
	SignErr    error
	VerifyErr  error
}

func (c *mockCrypto) Sign(_ []byte, _ string) ([]byte, error) {
	return c.SignResult, c.SignErr
}

func (c *mockCrypto) Verify(signature, msg []byte, keyID string) error {
	return c.VerifyErr
}

type mocGoAPIInteraction struct {
	GetQueryResult           *presexch.PresentationDefinition
	GetQueryError            error
	PresentCredentialErr     error
	VerifierDisplayDataRes   *openid4vp.VerifierDisplayData
	VerifierDisplayDataError error
}

func (o *mocGoAPIInteraction) GetQuery() (*presexch.PresentationDefinition, error) {
	return o.GetQueryResult, o.GetQueryError
}

func (o *mocGoAPIInteraction) PresentCredential(credentials []*afgoverifiable.Credential) error {
	return o.PresentCredentialErr
}

func (o *mocGoAPIInteraction) VerifierDisplayData() (*openid4vp.VerifierDisplayData, error) {
	return o.VerifierDisplayDataRes, o.VerifierDisplayDataError
}

type mocksDIDResolver struct {
	ResolveDocBytes []byte
	ResolveErr      error
}

func (m *mocksDIDResolver) Resolve(string) ([]byte, error) {
	return m.ResolveDocBytes, m.ResolveErr
}

func mockResolution(t *testing.T, vm *api.VerificationMethod) []byte {
	t.Helper()

	var (
		mockVM   *did.VerificationMethod
		err      error
		mockDID  string
		mockVMID string
	)

	idSplit := strings.Split(vm.ID, "#")
	switch len(idSplit) {
	case 1:
		mockVMID = idSplit[0]
	case 2:
		mockDID = idSplit[0]
		mockVMID = idSplit[1]
	default:
		t.Fail()
	}

	if vm.Key.JSONWebKey != nil {
		mockVM, err = did.NewVerificationMethodFromJWK(mockVMID, vm.Type, mockDID, vm.Key.JSONWebKey)
		require.NoError(t, err)
	} else {
		mockVM = did.NewVerificationMethodFromBytes(mockVMID, vm.Type, mockDID, vm.Key.Raw)
	}

	docRes := &did.DocResolution{
		DIDDocument: &did.Doc{
			ID:      mockDID,
			Context: []string{did.ContextV1},
			VerificationMethod: []did.VerificationMethod{
				*mockVM,
			},
			AssertionMethod: []did.Verification{
				{
					VerificationMethod: *mockVM,
				},
			},
		},
	}

	docBytes, err := docRes.JSONBytes()
	require.NoError(t, err)

	return docBytes
}
