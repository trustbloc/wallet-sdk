/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package wrapper contains wrappers that convert between the Go and gomobile APIs.
package wrapper

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

// MobileActivityLoggerWrapper is a goapi.ActivityLogger implementation that intercepts Activity log events, converts
// them to the gomobile-compatible Activity objects, and sends them to the wrapped mobile ActivityLogger.
type MobileActivityLoggerWrapper struct {
	MobileAPIActivityLogger api.ActivityLogger
}

// Log converts the given activity from a goapi.Activity object to a gomobile-compatible Activity object and then
// passes that converted object to the underlying mobile activity logger implementation.
func (m *MobileActivityLoggerWrapper) Log(activity *goapi.Activity) error {
	mobileActivity := &api.Activity{GoAPIActivity: activity}

	return m.MobileAPIActivityLogger.Log(mobileActivity)
}
