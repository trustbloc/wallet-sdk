/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package mem contains a simple in-memory api.ActivityLogger implementation.
package mem

import (
	"sync"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// ActivityLogger is a simple in-memory api.ActivityLogger implementation.
// New activities are appended to the array in the order that they are received.
type ActivityLogger struct {
	activities []*api.Activity
	lock       sync.RWMutex
}

// NewActivityLogger returns a new ActivityLogger.
func NewActivityLogger() *ActivityLogger {
	return &ActivityLogger{}
}

// Log logs a single activity.
// The activity is appended to the end of an internal array of activities.
func (m *ActivityLogger) Log(activity *api.Activity) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.activities = append(m.activities, activity)

	return nil
}

// Length returns the number of activities that this activity logger has logged.
func (m *ActivityLogger) Length() int {
	return len(m.activities)
}

// AtIndex returns the activity stored at the given index.
// Passing in 0 for the index will return the first activity in the underlying array. If this activity logger is only
// being used in one thread at a time, then that activity will be first activity that this activity logger has logged,
// and using higher indices will return more recent activities.
// If the index passed in is out of bounds, then nil is returned.
func (m *ActivityLogger) AtIndex(index int) *api.Activity {
	maxIndex := len(m.activities) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return m.activities[index]
}
