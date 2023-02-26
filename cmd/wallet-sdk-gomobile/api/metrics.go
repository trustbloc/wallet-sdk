/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import goapi "github.com/trustbloc/wallet-sdk/pkg/api"

// A MetricsEvent represents a single event that occurred that had performance metrics taken.
type MetricsEvent struct {
	// GoAPIMetricsEvent will not be accessible directly in the bindings (will be "skipped").
	// This is OK - this field is not intended to be set directly by the user of the SDK.
	GoAPIMetricsEvent *goapi.MetricsEvent
}

// Event returns the event name.
func (m *MetricsEvent) Event() string {
	return m.GoAPIMetricsEvent.Event
}

// ParentEvent returns the parent event name.
func (m *MetricsEvent) ParentEvent() string {
	return m.GoAPIMetricsEvent.ParentEvent
}

// DurationNanoseconds returns event duration as nanoseconds.
func (m *MetricsEvent) DurationNanoseconds() int64 {
	return m.GoAPIMetricsEvent.Duration.Nanoseconds()
}

// MetricsLogger represents a type that can log MetricsEvents.
type MetricsLogger interface {
	Log(metricsEvent *MetricsEvent) error
}
