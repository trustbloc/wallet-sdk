/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package testsupport mocks the trust registry API.
package testsupport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/trustregistry"
)

// ParseEvaluateIssuanceRequest parse evaluate issuance request to the trust registry.
func ParseEvaluateIssuanceRequest(r *http.Request) (*trustregistry.IssuanceRequest, error) {
	issuanceRequest := &trustregistry.IssuanceRequest{}

	err := json.NewDecoder(r.Body).Decode(&issuanceRequest)
	if err != nil {
		return nil, fmt.Errorf("fail to decode IssuanceRequest: %w", err)
	}

	return issuanceRequest, nil
}

// ParseEvaluatePresentationRequest parse evaluate presentation request to the trust registry.
func ParseEvaluatePresentationRequest(r *http.Request) (*trustregistry.PresentationRequest, error) {
	presentationRequest := &trustregistry.PresentationRequest{}

	err := json.NewDecoder(r.Body).Decode(&presentationRequest)
	if err != nil {
		return nil, fmt.Errorf("fail to decode PresentationRequest: %w", err)
	}

	return presentationRequest, nil
}

// HandleEvaluateIssuanceRequest mocks evaluate issuance API of the trust registry.
func HandleEvaluateIssuanceRequest(w http.ResponseWriter, r *http.Request) {
	issuanceRequest, err := ParseEvaluateIssuanceRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if issuanceRequest.IssuerDID == "did:web:correct.com" {
		writeResponse(w, &trustregistry.EvaluationResult{Allowed: true})

		return
	}

	if issuanceRequest.IssuerDID == "did:web:forbidden.com" {
		writeResponse(w, &trustregistry.EvaluationResult{
			ErrorCode:    "didForbidden",
			ErrorMessage: "Interaction with given issuer is forbidden",
		})

		return
	}

	writeResponse(w, &trustregistry.EvaluationResult{
		ErrorCode:    "didUnknown",
		ErrorMessage: "Issuer with given did is unknown",
	})
}

// HandleEvaluatePresentationRequest mocks evaluate presentation API of the trust registry.
func HandleEvaluatePresentationRequest(w http.ResponseWriter, r *http.Request) {
	presentationRequest, err := ParseEvaluatePresentationRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if presentationRequest.VerifierDid == "did:web:correct.com" {
		writeResponse(w, &trustregistry.EvaluationResult{Allowed: true})

		return
	}

	if presentationRequest.VerifierDid == "did:web:forbidden.com" {
		writeResponse(w, &trustregistry.EvaluationResult{
			ErrorCode:    "didForbidden",
			ErrorMessage: "Interaction with given verifier is forbidden",
		})

		return
	}

	writeResponse(w, &trustregistry.EvaluationResult{
		ErrorCode:    "didUnknown",
		ErrorMessage: "Verifier with given did is unknown",
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
