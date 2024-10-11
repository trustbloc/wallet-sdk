/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/trustbloc/vc-go/proof/defaults"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didion"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didjwk"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didkey"
	"github.com/trustbloc/wallet-sdk/pkg/common"

	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"
	diddoc "github.com/trustbloc/did-go/doc/did"
	afgotime "github.com/trustbloc/did-go/doc/util/time"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
	sdkapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
)

func TestCredentialAPI(t *testing.T) {
	kms, e := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, e)

	crypto := kms.GetCrypto()

	credStore := credential.NewInMemoryDB()

	didResolver, e := did.NewResolver(nil)
	require.NoError(t, e)

	signer := credential.NewSigner(didResolver, crypto)

	sdkResolver, e := resolver.NewDIDResolver()
	require.NoError(t, e)

	ldLoader := testutil.DocumentLoader(t)

	verifier := jwtvcVerifier{
		ldLoader: ldLoader,
		proofChecker: defaults.NewDefaultProofChecker(
			common.NewVDRKeyResolver(sdkResolver)),
	}

	testCases := []struct {
		name      string
		didMethod string
	}{
		{
			name:      "did:ion signing DID",
			didMethod: "ion",
		},
		{
			name:      "did:ion signing DID, with stored credential",
			didMethod: "ion",
		},
		{
			name:      "did:key signing DID",
			didMethod: "key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var didDoc *api.DIDDocResolution

			switch tc.didMethod {
			case "key":
				jwk, err := kms.Create(localkms.KeyTypeED25519)
				require.NoError(t, err)

				didDoc, err = didkey.Create(jwk)
				require.NoError(t, err)
			case "jwk":
				jwk, err := kms.Create(localkms.KeyTypeED25519)
				require.NoError(t, err)

				didDoc, err = didjwk.Create(jwk)
				require.NoError(t, err)
			case "ion":
				jwk, err := kms.Create(localkms.KeyTypeED25519)
				require.NoError(t, err)

				didDoc, err = didion.CreateLongForm(jwk)
				require.NoError(t, err)
			default:
				require.Fail(t, fmt.Sprintf("%s is not a supported DID method", tc.didMethod))
			}

			docID, err := didDoc.ID()
			require.NoError(t, err)

			templateCredential, err := afgoverifiable.CreateCredential(afgoverifiable.CredentialContents{
				ID:      "cred-ID",
				Types:   []string{afgoverifiable.VCType},
				Context: []string{afgoverifiable.V1ContextURI},
				Subject: []afgoverifiable.Subject{{
					ID: "foo",
				}},
				Issuer: &afgoverifiable.Issuer{
					ID: docID,
				},
				Issued: afgotime.NewTime(time.Now()),
			}, nil)
			require.NoError(t, err)

			err = credStore.Add(verifiable.NewCredential(templateCredential))
			require.NoError(t, err)

			cred := verifiable.NewCredential(templateCredential)

			issuedCred, err := signer.Issue(cred, docID)
			require.NoError(t, err)

			serializedCred, err := issuedCred.Serialize()
			require.NoError(t, err)

			require.NoError(t, verifier.verify([]byte(serializedCred)))
		})
	}
}

type jwtvcVerifier struct {
	ldLoader     ld.DocumentLoader
	proofChecker afgoverifiable.CombinedProofChecker
}

func (j *jwtvcVerifier) verify(cred []byte) error {
	_, err := afgoverifiable.ParseCredential(
		cred,
		afgoverifiable.WithJSONLDDocumentLoader(j.ldLoader),
		afgoverifiable.WithProofChecker(j.proofChecker),
	)

	return err
}

type didResolverWrapper struct {
	didResolver sdkapi.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}

type documentLoaderReverseWrapper struct {
	DocumentLoader ld.DocumentLoader
}

func (l *documentLoaderReverseWrapper) LoadDocument(url string) (*api.LDDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(url)
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
