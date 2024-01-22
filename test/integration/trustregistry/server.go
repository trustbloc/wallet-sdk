package main

import (
	"encoding/json"
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
	ForbiddenDIDs []string `json:"forbiddenDIDs"`
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
		writeResponse(w, &trustregistry.EvaluationResult{Allowed: true})

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

	if !slices.Contains(s.rules.ForbiddenDIDs, presentationRequest.VerifierDid) {
		writeResponse(w, &trustregistry.EvaluationResult{Allowed: true})

		return
	}

	writeResponse(w, &trustregistry.EvaluationResult{
		ErrorCode:    "didForbidden",
		ErrorMessage: "Interaction with given issuer is forbidden",
	})
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
