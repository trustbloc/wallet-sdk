/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package trustregistry implements trust registry API.
package trustregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	noopmetricslogger "github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

const (
	evaluateIssuanceEventText               = "Evaluate issuance"
	evaluateIssuanceEventVIAGetReqEventText = "Evaluate issuance via an HTTP GET request to %s"

	evaluatePresentationEventText               = "Evaluate presentation"
	evaluatePresentationEventVIAGetReqEventText = "Evaluate presentation via an HTTP GET request to %s"
)

// RegistryConfig is config for trust registry API.
type RegistryConfig struct {
	HTTPClient              *http.Client
	EvaluateIssuanceURL     string
	EvaluatePresentationURL string
	MetricsLogger           api.MetricsLogger
}

// Registry implements API for trust registry.
type Registry struct {
	httpClient              *http.Client
	evaluateIssuanceURL     string
	evaluatePresentationURL string
	metricsLogger           api.MetricsLogger
}

// New creates new trust registry API.
func New(config *RegistryConfig) *Registry {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	metricsLogger := config.MetricsLogger
	if metricsLogger == nil {
		metricsLogger = noopmetricslogger.NewMetricsLogger()
	}

	return &Registry{
		httpClient:              httpClient,
		evaluateIssuanceURL:     config.EvaluateIssuanceURL,
		evaluatePresentationURL: config.EvaluatePresentationURL,
		metricsLogger:           metricsLogger,
	}
}

// EvaluateIssuance evaluate is issuance request by calling trust registry.
func (r *Registry) EvaluateIssuance(request *IssuanceRequest) (*EvaluationResult, error) { //nolint:dupl
	req := httprequest.New(r.httpClient, r.metricsLogger)

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal issuanceRequest: %w", err)
	}

	responseBytes, err := req.Do(http.MethodPost, r.evaluateIssuanceURL, "application/json",
		bytes.NewBuffer(requestBytes),
		fmt.Sprintf(evaluateIssuanceEventVIAGetReqEventText, r.evaluateIssuanceURL),
		evaluateIssuanceEventText, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("evaluate issuance endpoint: %w", err)
	}

	evaluationResult := &EvaluationResult{}

	err = json.Unmarshal(responseBytes, evaluationResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal evaluate issuance endpoint: %w", err)
	}

	return evaluationResult, nil
}

// EvaluatePresentation evaluate is presentation request by calling trust registry.
func (r *Registry) EvaluatePresentation(request *PresentationRequest) (*EvaluationResult, error) { //nolint:dupl
	req := httprequest.New(r.httpClient, r.metricsLogger)

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal presentationRequest: %w", err)
	}

	responseBytes, err := req.Do(http.MethodPost, r.evaluatePresentationURL, "application/json",
		bytes.NewBuffer(requestBytes),
		fmt.Sprintf(evaluatePresentationEventVIAGetReqEventText, r.evaluatePresentationURL),
		evaluatePresentationEventText, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("evaluate presentation endpoint: %w", err)
	}

	evaluationResult := &EvaluationResult{}

	err = json.Unmarshal(responseBytes, evaluationResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal evaluate presentation endpoint: %w", err)
	}

	return evaluationResult, nil
}
