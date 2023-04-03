/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package otel contains functionality for open telemetry.
package otel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

const (
	supportedVersion  = "00"
	flags             = "01"
	traceparentHeader = "traceparent"
	traceIDSize       = 16
	spanIDSize        = 8
)

// Trace is a tool to simplify usage of open telemetry.
type Trace struct {
	traceID     string
	traceHeader *api.Header
}

// NewTrace returns new trace.
func NewTrace() (*Trace, error) {
	return GenerateTrace(rand.Read)
}

// GenerateTrace generates traces using provided random function.
func GenerateTrace(rndFn func(b []byte) (n int, err error)) (*Trace, error) {
	traceID, err := randomHex(traceIDSize, rndFn)
	if err != nil {
		return nil, fmt.Errorf("fail to create new trace: %w", err)
	}

	spanID, err := randomHex(spanIDSize, rndFn)
	if err != nil {
		return nil, fmt.Errorf("fail to create new trace: %w", err)
	}

	headerVal := fmt.Sprintf("%s-%s-%s-%s",
		supportedVersion,
		traceID,
		spanID,
		flags)

	return &Trace{
		traceID:     traceID,
		traceHeader: api.NewHeader(traceparentHeader, headerVal),
	}, nil
}

// TraceID returns id of the trace root.
func (t *Trace) TraceID() string {
	return t.traceID
}

// TraceHeader returns the trace HTTP header, which should be sent with an HTTP request to enable open telemetry.
func (t *Trace) TraceHeader() *api.Header {
	return t.traceHeader
}

func randomHex(n int, rndFn func(b []byte) (int, error)) (string, error) {
	bytes := make([]byte, n)
	if _, err := rndFn(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
