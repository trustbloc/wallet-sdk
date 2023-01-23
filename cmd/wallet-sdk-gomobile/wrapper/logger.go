/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"encoding/json"
	"fmt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

// MobileLoggerWrapper is a goapi.Logger implementation that acts as a conversion layer between the goapi Logger
// interface and a mobile api.Logger implementation.
type MobileLoggerWrapper struct {
	MobileAPILogger api.Logger
}

// Log marshals the given log entry into JSON and passes it to the underlying mobile API logger implementation.
func (m *MobileLoggerWrapper) Log(log *goapi.LogEntry) {
	marshalledLog, err := json.Marshal(log)
	if err != nil {
		println(fmt.Sprintf("failed to write a log message (type:%s)", log.Type))
	}

	m.MobileAPILogger.Log(string(marshalledLog))
}
