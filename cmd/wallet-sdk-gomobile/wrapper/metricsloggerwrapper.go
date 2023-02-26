/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	gomobileapi "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

// MobileMetricsLoggerWrapper is a goapi.MetricsLogger implementation that intercepts metrics log events, converts
// them to the gomobile-compatible MetricsEvent objects, and sends them to the wrapped mobile MetricsLogger.
type MobileMetricsLoggerWrapper struct {
	MobileAPIMetricsLogger gomobileapi.MetricsLogger
}

// Log converts the given activity from a goapi.MetricsEvent object to a gomobile-compatible MetricsEvent object and
// then passes that converted object to the underlying MetricsLogger implementation.
func (m *MobileMetricsLoggerWrapper) Log(metricsEvent *goapi.MetricsEvent) error {
	if m.MobileAPIMetricsLogger == nil {
		return nil
	}

	mobileMetricsEvent := &gomobileapi.MetricsEvent{GoAPIMetricsEvent: metricsEvent}

	return m.MobileAPIMetricsLogger.Log(mobileMetricsEvent)
}
