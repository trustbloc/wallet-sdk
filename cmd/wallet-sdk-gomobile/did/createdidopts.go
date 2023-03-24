/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// CreateOpts contains all optional arguments that can be passed into the CreateDID method.
type CreateOpts struct {
	verificationType string
	keyType          string
	metricsLogger    api.MetricsLogger
}

// NewCreateOpts returns a new CreateOpts object.
func NewCreateOpts() *CreateOpts {
	return &CreateOpts{}
}

// SetVerificationType sets a verification type to use.
func (c *CreateOpts) SetVerificationType(verificationType string) {
	c.verificationType = verificationType
}

// SetKeyType sets the key type to use for keys generated during DID creation.
// This option is ignored for DID:ion's update and recovery key types (they're hardcoded to an ECDSA P-256 type).
func (c *CreateOpts) SetKeyType(keyType string) {
	c.keyType = keyType
}

// SetMetricsLogger sets a metrics logger to use.
func (c *CreateOpts) SetMetricsLogger(metricsLogger api.MetricsLogger) {
	c.metricsLogger = metricsLogger
}
