/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

// Constructors workaround to generate swift compatible objective-c functions.
type Constructors struct{}

// NewKMS workaround to generate swift compatible objective-c functions.
func (c *Constructors) NewKMS(kmsStore Store) (*KMS, error) {
	return NewKMS(kmsStore)
}
