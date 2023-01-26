/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

// Data represents specific details of an activity.
type Data struct {
	Client    string      `json:"client,omitempty"`
	Operation string      `json:"operation,omitempty"`
	Status    string      `json:"status,omitempty"`
	Params    *JSONObject `json:"params,omitempty"`
}

// An Activity represents a single activity.
type Activity struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Time int64  `json:"time,omitempty"` // Unix time
	Data *Data  `json:"data,omitempty"`
}

// An ActivityLogger logs activities.
type ActivityLogger interface {
	// Log logs a single activity.
	Log(activity *Activity) error
}
