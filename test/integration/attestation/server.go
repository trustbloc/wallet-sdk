/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	tlsutils "github.com/trustbloc/cmdutil-go/pkg/utils/tls"
	"github.com/trustbloc/did-go/doc/did"
	utiltime "github.com/trustbloc/did-go/doc/util/time"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/proof/testsupport"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	didion "github.com/trustbloc/wallet-sdk/pkg/did/creator/ion"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

type sessionMetadata struct {
	challenge string
	payload   map[string]interface{}
}

type serverConfig struct {
}

type server struct {
	router        *mux.Router
	httpClient    *http.Client
	sessions      sync.Map // sessionID -> sessionMetadata
	config        *serverConfig
	cryptoProfile *cryptoProfile
}

func newServer(config *serverConfig) *server {
	router := mux.NewRouter()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	customCAs := os.Getenv("ROOT_CA_CERTS_PATH")

	if customCAs != "" {
		rootCAs, err := tlsutils.GetCertPool(false, []string{customCAs})
		if err != nil {
			panic(err)
		}

		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            rootCAs,
			MinVersion:         tls.VersionTLS12,
		}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	cryptoProfile, err := createDID()
	if err != nil {
		log.Fatalf("ATTESTATION_PROFILE is required")
	}

	srv := &server{
		router:        router,
		httpClient:    httpClient,
		config:        config,
		cryptoProfile: cryptoProfile,
	}

	router.HandleFunc("/profiles/profileID/profileVersion/wallet/attestation/init", srv.evaluateWalletAttestationInitRequest).Methods(http.MethodPost)
	router.HandleFunc("/profiles/profileID/profileVersion/wallet/attestation/complete", srv.evaluateWalletAttestationCompleteRequest).Methods(http.MethodPost)

	return srv
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) evaluateWalletAttestationInitRequest(w http.ResponseWriter, r *http.Request) {
	var request AttestWalletInitRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		s.writeResponse(
			w, http.StatusBadRequest, fmt.Sprintf("decode wallet attestation init request: %s", err.Error()))

		return
	}

	authorization := r.Header.Get("Authorization")
	if authorization != "Bearer token" {
		s.writeResponse(w, http.StatusUnauthorized, fmt.Sprintf("authorization header is invalid: %q", authorization))

		return
	}

	log.Printf("handling request: %s with payload %v", r.URL.String(), request)

	sessionID, challenge := uuid.NewString(), uuid.NewString()

	response := &AttestWalletInitResponse{
		Challenge: challenge,
		SessionID: sessionID,
	}

	s.sessions.Store(sessionID, sessionMetadata{
		challenge: challenge,
		payload:   request.Payload,
	})

	go func() {
		time.Sleep(5 * time.Minute)
		s.sessions.Delete(sessionID)

		log.Printf("session %s is deleted", sessionID)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("failed to write response: %s", err.Error())
	}
}

func (s *server) evaluateWalletAttestationCompleteRequest(w http.ResponseWriter, r *http.Request) {
	var request AttestWalletCompleteRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		s.writeResponse(
			w, http.StatusBadRequest, fmt.Sprintf("decode wallet attestation init request: %s", err.Error()))

		return
	}

	authorization := r.Header.Get("Authorization")
	if authorization != "Bearer token" {
		s.writeResponse(w, http.StatusUnauthorized, fmt.Sprintf("authorization header is invalid: %q", authorization))

		return
	}

	log.Printf("handling request: %s with payload %v", r.URL.String(), request)

	if request.AssuranceLevel != "low" {
		s.writeResponse(w, http.StatusBadRequest, "assuranceLevel field is invalid")

		return
	}

	if request.Proof.ProofType != "jwt" {
		s.writeResponse(w, http.StatusBadRequest, "proof.ProofType field is invalid")

		return
	}

	walletDID, sesData, err := s.evaluateWalletProofJWT(request.SessionID, request.Proof.Jwt)
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	attestationVC, err := s.attestationVC(context.Background(), walletDID, sesData)
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, err.Error())

		return
	}

	response := &AttestWalletCompleteResponse{
		WalletAttestationVC: attestationVC,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("failed to write response: %s", err.Error())
	}
}

func (s *server) evaluateWalletProofJWT(
	sessionID, proofJWT string,
) (string, *sessionMetadata, error) {
	jwtParsed, _, err := jwt.Parse(proofJWT)
	if err != nil {
		return "", nil, fmt.Errorf("parse request.Proof.Jwt: %s", err.Error())
	}

	var jwtProofClaims JwtProofClaims
	err = jwtParsed.DecodeClaims(&jwtProofClaims)
	if err != nil {
		return "", nil, fmt.Errorf("decode request.Proof.Jwt: %s", err.Error())
	}

	var sessionData sessionMetadata
	sessionDataIface, ok := s.sessions.Load(sessionID)
	if ok {
		sessionData, ok = sessionDataIface.(sessionMetadata)
	}

	if !ok {
		return "", nil, fmt.Errorf("session %s is unknown", sessionID)
	}

	if jwtProofClaims.Audience == "" {
		return "", nil, fmt.Errorf("jwtProofClaims.Audience is empty")
	}

	if jwtProofClaims.IssuedAt == 0 {
		return "", nil, fmt.Errorf("jwtProofClaims.IssuedAt is invalid")
	}

	if jwtProofClaims.Exp == 0 {
		return "", nil, fmt.Errorf("jwtProofClaims.Exp is invalid")
	}

	if jwtProofClaims.Nonce != sessionData.challenge {
		return "", nil, fmt.Errorf("jwtProofClaims.Nonce is invalid, got: %s, want: %s", jwtProofClaims.Nonce, sessionData.challenge)
	}

	return jwtProofClaims.Issuer, &sessionData, nil
}

func (s *server) attestationVC(ctx context.Context, walletDID string, ses *sessionMetadata) (string, error) {
	vcc := verifiable.CredentialContents{
		Context: []string{
			verifiable.V1ContextURI,
			"https://w3c-ccg.github.io/lds-jws2020/contexts/lds-jws2020-v1.json",
		},
		ID: uuid.New().String(),
		Types: []string{
			verifiable.VCType,
			"WalletAttestationCredential",
		},
		Subject: []verifiable.Subject{
			{
				ID:           walletDID,
				CustomFields: ses.payload,
			},
		},
		Issuer: &verifiable.Issuer{
			ID: s.cryptoProfile.did.DIDDocument.ID,
		},
		Issued: &utiltime.TimeWrapper{
			Time: time.Now(),
		},
		Expired: &utiltime.TimeWrapper{
			Time: time.Now().Add(time.Hour),
		},
	}

	vc, err := verifiable.CreateCredential(vcc, nil)
	if err != nil {
		return "", fmt.Errorf("create attestation vc: %w", err)
	}

	claims, err := vc.JWTClaims(false)
	if err != nil {
		return "", fmt.Errorf("get jwt claims: %w", err)
	}

	jwsAlgo, err := verifiable.KeyTypeToJWSAlgo(kms.ED25519Type)
	if err != nil {
		return "", err
	}

	jws, err := claims.MarshalJWSString(jwsAlgo, testsupport.NewProofCreator(&signerWithEmbeddedKey{
		crypto: s.cryptoProfile.crypto,
		kid:    s.cryptoProfile.kid,
	}), s.cryptoProfile.did.DIDDocument.VerificationMethod[0].ID)
	if err != nil {
		return "", fmt.Errorf("marshal unsecured jwt: %w", err)
	}

	return jws, nil
}

// writeResponse writes interface value to response
func (s *server) writeResponse(
	rw http.ResponseWriter,
	status int,
	msg string,
) {
	log.Printf("[%d]   %s", status, msg)

	rw.WriteHeader(status)

	_, _ = rw.Write([]byte(msg))
}

type signerWithEmbeddedKey struct {
	crypto api.Crypto
	kid    string
}

func (s *signerWithEmbeddedKey) Sign(data []byte) ([]byte, error) {
	return s.crypto.Sign(data, s.kid)
}

type cryptoProfile struct {
	crypto api.Crypto
	kid    string
	did    *did.DocResolution
}

func createDID() (*cryptoProfile, error) {
	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: localkms.NewMemKMSStore()})
	if err != nil {
		return nil, fmt.Errorf("create new local KMS: %w", err)
	}

	kid, jwk, err := localKMS.Create(kms.ED25519Type)
	if err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	didDoc, err := didion.CreateLongForm(jwk)
	if err != nil {
		return nil, fmt.Errorf("create did: %w", err)
	}

	return &cryptoProfile{
		crypto: localKMS.GetCrypto(),
		kid:    kid,
		did:    didDoc,
	}, nil
}
