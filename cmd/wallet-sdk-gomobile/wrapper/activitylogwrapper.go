/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"encoding/json"

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
	mobileActivity := &api.Activity{
		ID:   activity.ID.String(),
		Type: activity.Type,
		Time: activity.Time.Unix(),
		Data: &api.Data{
			Client:    activity.Data.Client,
			Operation: activity.Data.Operation,
			Status:    activity.Data.Status,
		},
	}

	if activity.Data.Params != nil {
		marshalledParams, err := json.Marshal(activity.Data.Params)
		if err != nil {
			return err
		}

		mobileActivity.Data.Params = &api.JSONObject{Data: marshalledParams}
	}

	return m.MobileAPIActivityLogger.Log(mobileActivity)
}
