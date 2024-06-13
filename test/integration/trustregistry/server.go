package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"

	"github.com/trustbloc/wallet-sdk/pkg/trustregistry"
	"github.com/trustbloc/wallet-sdk/pkg/trustregistry/testsupport"
)

type Server struct {
	router *mux.Router
	rules  *Rules
}

type Config struct {
	rulesJSONPath string
}

type Rules struct {
	ForbiddenDIDs  []string `json:"forbiddenDIDs"`
	ForbiddenTypes []string `json:"forbiddenTypes"`
}

func NewServer(c *Config) (*Server, error) {
	router := mux.NewRouter()

	rules, err := readRules(c.rulesJSONPath)
	if err != nil {
		return nil, err
	}

	server := &Server{
		router: router,
		rules:  rules,
	}

	router.HandleFunc("/wallet/interactions/issuance", server.handleEvaluateIssuanceRequest).Methods(http.MethodPost)
	router.HandleFunc("/wallet/interactions/presentation", server.handleEvaluatePresentationRequest).Methods(http.MethodPost)

	router.HandleFunc("/issuer/policies/{policyID}/{policyVersion}/interactions/issuance", server.evaluateIssuerIssuancePolicy).Methods(http.MethodPost)
	router.HandleFunc("/verifier/policies/{policyID}/{policyVersion}/interactions/presentation", server.evaluateVerifierPresentationPolicy).Methods(http.MethodPost)

	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// HandleEvaluateIssuanceRequest mocks evaluate issuance API of the trust registry.
func (s *Server) handleEvaluateIssuanceRequest(w http.ResponseWriter, r *http.Request) {
	issuanceRequest, err := testsupport.ParseEvaluateIssuanceRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if !slices.Contains(s.rules.ForbiddenDIDs, issuanceRequest.IssuerDID) {
		writeResponse(w, &trustregistry.EvaluationResult{
			Allowed: true,
			Data: &trustregistry.EvaluationData{
				AttestationsRequired: []string{"wallet_authentication"},
			},
		})

		return
	}

	writeResponse(w, &trustregistry.EvaluationResult{
		ErrorCode:    "didForbidden",
		ErrorMessage: "Interaction with given issuer is forbidden",
	})
}

// HandleEvaluatePresentationRequest mocks evaluate presentation API of the trust registry.
func (s *Server) handleEvaluatePresentationRequest(w http.ResponseWriter, r *http.Request) {
	presentationRequest, err := testsupport.ParseEvaluatePresentationRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	for _, cred := range presentationRequest.CredentialClaims {
		fmt.Printf("credential claims: %+v\n", cred.CredentialClaimKeys)

		if cred.CredentialClaimKeys == nil {
			writeResponse(w, &trustregistry.EvaluationResult{
				ErrorCode:    "claimsForbidden",
				ErrorMessage: "Interaction without credential_claim_keys is forbidden",
			})

			return
		}
		for _, credType := range cred.CredentialTypes {
			if slices.Contains(s.rules.ForbiddenTypes, credType) {
				writeResponse(w, &trustregistry.EvaluationResult{
					ErrorCode:    "typeForbidden",
					ErrorMessage: fmt.Sprintf("Interaction with given type %q is forbidden", credType),
				})

				return
			}
		}
	}

	if slices.Contains(s.rules.ForbiddenDIDs, presentationRequest.VerifierDid) {
		writeResponse(w, &trustregistry.EvaluationResult{
			ErrorCode:    "didForbidden",
			ErrorMessage: "Interaction with given issuer is forbidden",
		})
		return
	}

	writeResponse(w, &trustregistry.EvaluationResult{
		Allowed: true,
		Data: &trustregistry.EvaluationData{
			AttestationsRequired: []string{"wallet_authentication"},
		},
	})
}

func (s *Server) evaluateIssuerIssuancePolicy(w http.ResponseWriter, r *http.Request) {
	var request IssuerIssuanceRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		writeResponseWithStatus(
			w, http.StatusBadRequest+5, fmt.Sprintf("decode issuance policy request: %s", err.Error()))

		return
	}

	if request.IssuerDID == "" {
		writeResponseWithStatus(
			w, http.StatusBadRequest+6, "issuer did is empty")

		return
	}

	if request.AttestationVC == nil || len(*request.AttestationVC) == 0 {
		writeResponseWithStatus(
			w, http.StatusBadRequest+7, "no attestation vc supplied")

		return
	}

	log.Printf("handling request: %s with payload %v", r.URL.String(), request)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := &PolicyEvaluationResponse{
		Allowed: true,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("failed to write response: %s", err.Error())
	}
}

func (s *Server) evaluateVerifierPresentationPolicy(w http.ResponseWriter, r *http.Request) {
	var request VerifierPresentationRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		writeResponseWithStatus(
			w, http.StatusBadRequest, fmt.Sprintf("decode presentation policy request: %s", err.Error()))

		return
	}

	if request.VerifierDID == "" {
		writeResponseWithStatus(
			w, http.StatusBadRequest, "verifier did is empty")

		return
	}
	if request.CredentialMatches == nil {
		request.CredentialMatches = request.CredentialMetadata
	}

	if request.CredentialMatches == nil || len(request.CredentialMatches) == 0 {
		writeResponseWithStatus(
			w, http.StatusBadRequest, "no credential matches supplied")

		return
	}

	if request.AttestationVC == nil || len(*request.AttestationVC) == 0 {
		writeResponseWithStatus(
			w, http.StatusBadRequest, "no attestation vc supplied")

		return
	}

	log.Printf("handling request: %s with payload %v", r.URL.String(), request)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := &PolicyEvaluationResponse{
		Allowed: true,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("failed to write response: %s", err.Error())
	}
}

func writeResponse(w http.ResponseWriter, body interface{}) {
	bytes, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Write failed: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

// writeResponse writes interface value to response
func writeResponseWithStatus(
	rw http.ResponseWriter,
	status int,
	msg string,
) {
	log.Printf("[%d]   %s", status, msg)

	rw.WriteHeader(status)

	_, _ = rw.Write([]byte(msg))
}

func readRules(filePath string) (*Rules, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var rules Rules
	err = json.Unmarshal(byteValue, &rules)
	if err != nil {
		return nil, err
	}
	return &rules, nil
}
