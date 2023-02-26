/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "time"

// A MetricsEvent represents a single event that occurred that had performance metrics taken.
type MetricsEvent struct {
	// Event is the name of the event that occurred.
	Event string
	// ParentEvent is the name of the event that encompasses this event. Some longer operations log a larger event
	// that captures the overall time of the method, and during that method some sub-events are also logged.
	// If the parent event info is empty, then this MetricsEvent is a "root" event.
	// Sub-events always have a duration that is <= the duration of the parent event.
	ParentEvent string
	// Duration is how long the event/operation took.
	Duration time.Duration
}

// MetricsLogger represents a type that can log MetricsEvents.
type MetricsLogger interface {
	Log(event *MetricsEvent) error
}
