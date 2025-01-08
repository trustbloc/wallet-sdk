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
	"strings"
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	wrapperapi "github.com/trustbloc/kms-go/wrapper/api"
	"github.com/trustbloc/vc-go/presexch"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	gomobdid "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

var (
	//go:embed test_data/request_object.jwt
	requestObjectJWT string

	//go:embed test_data/credentials.jsonld
	credentialsJSONLD []byte
)

func TestNewInteraction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("OpenTelemetry disabled, custom headers used instead", func(t *testing.T) {
			resolver, err := gomobdid.NewResolver(gomobdid.NewResolverOpts())
			require.NoError(t, err)

			requiredArgs := NewArgs(
				requestObjectJWT,
				&mockCrypto{},
				resolver,
			)

			// Note: in-depth testing of opts functionality is done in the integration tests.
			opts := NewOpts()

			additionalHeaders := api.NewHeaders()

			additionalHeaders.Add(api.NewHeader("header-name-1", "header-value-1"))
			additionalHeaders.Add(api.NewHeader("header-name-2", "header-value-2"))

			opts.AddHeaders(additionalHeaders)

			opts.DisableOpenTelemetry()

			instance, err := NewInteraction(requiredArgs, opts)
			require.NoError(t, err)
			require.NotNil(t, instance)
			require.Equal(t, 0, instance.CustomScope().Length())
		})
		t.Run("All other options invoked", func(t *testing.T) {
			resolver, err := gomobdid.NewResolver(gomobdid.NewResolverOpts())
			require.NoError(t, err)

			requiredArgs := NewArgs(
				requestObjectJWT,
				&mockCrypto{},
				resolver,
			)

			// Note: in-depth testing of opts functionality is done in the integration tests.
			opts := NewOpts()
			opts.SetDocumentLoader(&documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)})
			opts.SetActivityLogger(mem.NewActivityLogger())
			opts.SetMetricsLogger(stderr.NewMetricsLogger())
			opts.DisableHTTPClientTLSVerify()
			opts.SetHTTPTimeoutNanoseconds(0)

			localKMS, err := localkms.NewKMS(localkms.NewMemKMSStore())
			require.NoError(t, err)

			opts.EnableAddingDIProofs(localKMS)

			instance, err := NewInteraction(requiredArgs, opts)
			require.NoError(t, err)
			require.NotNil(t, instance)
		})
	})

	t.Run("Failure - invalid authorization request", func(t *testing.T) {
		requiredArgs := NewArgs(
			requestObjectJWT,
			&mockCrypto{},
			&mockDIDResolver{},
		)

		instance, err := NewInteraction(requiredArgs, nil)
		testutil.RequireErrorContains(t, err, "INVALID_AUTHORIZATION_REQUEST")
		testutil.RequireErrorContains(t, err, "verify request object: check proof: invalid public key id:"+
			" resolve DID did:ion:EiDYWcDuP-EDjVyFWGFdpgPncar9A7OGFykdeX71ZTU-wg:")
		require.Nil(t, instance)
	})
}

func TestOpenID4VP_PresentCredential(t *testing.T) {
	mockKey, _, e := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, e)

	credentials := verifiable.NewCredentialsArray()

	var credentialData []json.RawMessage

	e = json.Unmarshal(credentialsJSONLD, &credentialData)
	require.NoError(t, e)

	for _, credBytes := range credentialData {
		cred, err := afgoverifiable.ParseCredential(credBytes,
			afgoverifiable.WithDisabledProofCheck(), afgoverifiable.WithCredDisableValidation())
		require.NoError(t, err)

		credentials.Add(verifiable.NewCredential(cred))
	}

	singleCredential := credentials.AtIndex(0)

	verificationMethod := &api.VerificationMethod{
		ID:   "did:example:12345#testId",
		Type: "Ed25519VerificationKey2018",
		Key:  models.VerificationKey{Raw: mockKey},
	}

	makeInteraction := func() *Interaction {
		return &Interaction{
			crypto:           &mockCrypto{},
			ldDocumentLoader: &documentLoaderWrapper{goAPIDocumentLoader: testutil.DocumentLoader(t)},
			goAPIOpenID4VP: &mockGoAPIInteraction{
				PresentCredentialErr: nil,
			},
			didResolver: &mockDIDResolver{ResolveDocBytes: mockResolution(t, verificationMethod)},
		}
	}

	t.Run("Success", func(t *testing.T) {
		instance := makeInteraction()

		query, err := instance.GetQuery()
		require.NotNil(t, query)
		require.NoError(t, err)

		err = instance.PresentCredential(credentials)
		require.NoError(t, err)
	})

	t.Run("Success With Opts", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredentialOpts(credentials, NewPresentCredentialOpts().
			AddScopeClaim("claim1", `{"key" : "val"}`).
			SetAttestationVC(verificationMethod, "invalidVC").
			SetInteractionDetails(`{"key1": "value1"}`))
		require.NoError(t, err)
	})

	t.Run("Success With nil Opts", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredentialOpts(credentials, nil)
		require.NoError(t, err)
	})

	t.Run("Success Unsafe", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredentialUnsafe(singleCredential)
		require.NoError(t, err)
	})

	t.Run("Present credentials failed", func(t *testing.T) {
		instance := makeInteraction()

		instance.goAPIOpenID4VP = &mockGoAPIInteraction{
			PresentCredentialErr: errors.New("present credentials failed"),
		}

		err := instance.PresentCredential(credentials)
		require.Contains(t, err.Error(), "present credentials failed")
	})

	t.Run("Present credentials with opts failed", func(t *testing.T) {
		instance := makeInteraction()

		instance.goAPIOpenID4VP = &mockGoAPIInteraction{
			PresentCredentialErr: errors.New("present credentials failed"),
		}

		err := instance.PresentCredentialOpts(credentials, NewPresentCredentialOpts().
			AddScopeClaim("claim1", `{"key" : "val"}`))
		require.Contains(t, err.Error(), "present credentials failed")
	})

	t.Run("Present credentials with invalid scope value", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredentialOpts(credentials, NewPresentCredentialOpts().
			AddScopeClaim("claim1", `"key" : "val"`))
		require.ErrorContains(t, err, `fail to parse "claim1" claim json`)
	})

	t.Run("Present credentials with invalid interaction details", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredentialOpts(credentials, NewPresentCredentialOpts().
			SetInteractionDetails(`"key1": "value1"`))
		require.ErrorContains(t, err, `decode vp interaction details`)
	})

	t.Run("Present credentials unsafe failed", func(t *testing.T) {
		instance := makeInteraction()

		instance.goAPIOpenID4VP = &mockGoAPIInteraction{
			PresentCredentialUnsafeErr: errors.New("present credentials failed"),
		}

		err := instance.PresentCredentialUnsafe(singleCredential)
		require.Contains(t, err.Error(), "present credentials failed")
	})

	t.Run("CredentialsArray object is nil", func(t *testing.T) {
		instance := makeInteraction()

		err := instance.PresentCredential(nil)
		testutil.RequireErrorContains(t, err, "credentialsArray object cannot be nil")
	})

	t.Run("Nil Credential object in array", func(t *testing.T) {
		instance := makeInteraction()

		credentials.Add(nil)

		err := instance.PresentCredential(credentials)
		testutil.RequireErrorContains(t, err, "credential objects cannot be nil "+
			"(credential at index 5 is nil)")
	})
}

func TestGetCustomClaims(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		claims, err := getCustomClaims(NewPresentCredentialOpts().
			AddScopeClaim("claim1", `{"key" : "val"}`))
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{
			"claim1": map[string]interface{}{
				"key": "val",
			},
		},
			claims.ScopeClaims)
	})
}

func TestInteraction_PresentedClaims(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				PresentedClaimsResult: map[string]interface{}{
					"claim1": "val1",
				},
			},
		}

		claims, err := instance.PresentedClaims(&verifiable.Credential{})
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{"claim1": "val1"}, claims.ContentJSON)
	})

	t.Run("Failure", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				PresentedClaimsErr: errors.New("presented claims err"),
			},
		}

		_, err := instance.PresentedClaims(&verifiable.Credential{})
		require.ErrorContains(t, err, "presented claims err")
	})
}

func TestInteraction_VerifierDisplayData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				VerifierDisplayDataRes: &openid4vp.VerifierDisplayData{
					DID:     "DID",
					Name:    "testName",
					Purpose: "purpose",
					LogoURI: "logoURI",
				},
			},
		}

		data := instance.VerifierDisplayData()
		require.NotNil(t, data)
		require.Equal(t, "DID", data.DID())
		require.Equal(t, "testName", data.Name())
		require.Equal(t, "purpose", data.Purpose())
		require.Equal(t, "logoURI", data.LogoURI())
	})
}

func TestInteraction_TrustInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				VerifierTrustInfo: &openid4vp.VerifierTrustInfo{
					DID:    "TestDID",
					Domain: "TestDomain",
				},
			},
		}

		info, err := instance.TrustInfo()
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, "TestDID", info.DID)
		require.Equal(t, "TestDomain", info.Domain)
	})

	t.Run("Failure", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				VerifierTrustInfoErr: errors.New("trust info err"),
			},
		}

		_, err := instance.TrustInfo()
		require.ErrorContains(t, err, "trust info err")
	})
}

func TestInteraction_Acknowledgment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := &Interaction{
			goAPIOpenID4VP: &mockGoAPIInteraction{
				AcknowledgmentResult: &openid4vp.Acknowledgment{
					ResponseURI: "https://verifier/present",
					State:       "98822a39-9178-4742-a2dc-aba49879fc7b",
				},
			},
		}

		ack := instance.Acknowledgment()

		require.NotNil(t, ack)
		require.Equal(t, "https://verifier/present", ack.acknowledgment.ResponseURI)
		require.Equal(t, "98822a39-9178-4742-a2dc-aba49879fc7b", ack.acknowledgment.State)

		err := ack.SetInteractionDetails(`{"key1": "value1"}`)
		require.NoError(t, err)

		serialized, err := ack.Serialize()
		require.NoError(t, err)

		ackRestored, err := NewAcknowledgment(serialized)
		require.NoError(t, err)
		require.Equal(t, ack.acknowledgment.ResponseURI, ackRestored.acknowledgment.ResponseURI)
		require.Equal(t, ack.acknowledgment.State, ackRestored.acknowledgment.State)
		require.Equal(t, map[string]interface{}{"key1": "value1"}, ackRestored.acknowledgment.InteractionDetails)
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

func (c *mockCrypto) Verify(_, _ []byte, _ string) error {
	return c.VerifyErr
}

type mockGoAPIInteraction struct {
	GetQueryResult             *presexch.PresentationDefinition
	ScopeResult                []string
	PresentCredentialErr       error
	PresentCredentialUnsafeErr error
	VerifierDisplayDataRes     *openid4vp.VerifierDisplayData
	VerifierTrustInfo          *openid4vp.VerifierTrustInfo
	VerifierTrustInfoErr       error

	PresentedClaimsResult interface{}
	PresentedClaimsErr    error

	AcknowledgmentResult *openid4vp.Acknowledgment
}

func (o *mockGoAPIInteraction) GetQuery() *presexch.PresentationDefinition {
	return o.GetQueryResult
}

func (o *mockGoAPIInteraction) CustomScope() []string {
	return o.ScopeResult
}

func (o *mockGoAPIInteraction) PresentCredential(
	[]*afgoverifiable.Credential,
	openid4vp.CustomClaims,
	...openid4vp.PresentOpt,
) error {
	return o.PresentCredentialErr
}

func (o *mockGoAPIInteraction) PresentedClaims(*afgoverifiable.Credential) (interface{}, error) {
	return o.PresentedClaimsResult, o.PresentedClaimsErr
}

func (o *mockGoAPIInteraction) PresentCredentialUnsafe(*afgoverifiable.Credential, openid4vp.CustomClaims) error {
	return o.PresentCredentialUnsafeErr
}

func (o *mockGoAPIInteraction) VerifierDisplayData() *openid4vp.VerifierDisplayData {
	return o.VerifierDisplayDataRes
}

func (o *mockGoAPIInteraction) TrustInfo() (*openid4vp.VerifierTrustInfo, error) {
	return o.VerifierTrustInfo, o.VerifierTrustInfoErr
}

func (o *mockGoAPIInteraction) Acknowledgment() *openid4vp.Acknowledgment {
	return o.AcknowledgmentResult
}

type mockDIDResolver struct {
	ResolveDocBytes []byte
	ResolveErr      error
}

func (m *mockDIDResolver) Resolve(string) ([]byte, error) {
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

type mockSuite struct{}

var errSuiteNoSupport = errors.New("mock suite supports nothing")

func (m mockSuite) KeyCreator() (wrapperapi.KeyCreator, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) RawKeyCreator() (wrapperapi.RawKeyCreator, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) KMSCrypto() (wrapperapi.KMSCrypto, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) KMSCryptoSigner() (wrapperapi.KMSCryptoSigner, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) KMSCryptoMultiSigner() (wrapperapi.KMSCryptoMultiSigner, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) KMSCryptoVerifier() (wrapperapi.KMSCryptoVerifier, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) EncrypterDecrypter() (wrapperapi.EncrypterDecrypter, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) FixedKeyCrypto(*jwk.JWK) (wrapperapi.FixedKeyCrypto, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) FixedKeySigner(string) (wrapperapi.FixedKeySigner, error) {
	return nil, errSuiteNoSupport
}

func (m mockSuite) FixedKeyMultiSigner(string) (wrapperapi.FixedKeyMultiSigner, error) {
	return nil, errSuiteNoSupport
}

var _ wrapperapi.Suite = &mockSuite{}
