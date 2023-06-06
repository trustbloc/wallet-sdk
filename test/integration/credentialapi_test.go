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

	diddoc "github.com/hyperledger/aries-framework-go/component/models/did"
	afgotime "github.com/hyperledger/aries-framework-go/component/models/util/time"
	afgoverifiable "github.com/hyperledger/aries-framework-go/component/models/verifiable"
	vdrapi "github.com/hyperledger/aries-framework-go/spi/vdr"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"
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

	c, e := did.NewCreator(kms)
	require.NoError(t, e)

	sdkResolver, e := resolver.NewDIDResolver()
	require.NoError(t, e)

	ldLoader := testutil.DocumentLoader(t)

	verifier := jwtvcVerifier{
		ldLoader: ldLoader,
		publicKeyFetcher: afgoverifiable.NewVDRKeyResolver(&didResolverWrapper{
			didResolver: sdkResolver,
		}).PublicKeyFetcher(),
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
			createDIDOptionalArgs := did.NewCreateOpts()
			createDIDOptionalArgs.SetMetricsLogger(stderr.NewMetricsLogger())

			didDoc, err := c.Create(tc.didMethod, createDIDOptionalArgs)
			require.NoError(t, err)

			docID, err := didDoc.ID()
			require.NoError(t, err)

			templateCredential := &afgoverifiable.Credential{
				ID:      "cred-ID",
				Types:   []string{afgoverifiable.VCType},
				Context: []string{afgoverifiable.ContextURI},
				Subject: afgoverifiable.Subject{
					ID: "foo",
				},
				Issuer: afgoverifiable.Issuer{
					ID: docID,
				},
				Issued: afgotime.NewTime(time.Now()),
			}

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
	ldLoader         ld.DocumentLoader
	publicKeyFetcher afgoverifiable.PublicKeyFetcher
}

func (j *jwtvcVerifier) verify(cred []byte) error {
	_, err := afgoverifiable.ParseCredential(
		cred,
		afgoverifiable.WithJSONLDDocumentLoader(j.ldLoader),
		afgoverifiable.WithPublicKeyFetcher(j.publicKeyFetcher),
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
