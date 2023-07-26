/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package util implements varies utils used in wallet-sdk-js.
package util

// MapTo map array to another array using f function.
func MapTo[T, U any](ts []T, f func(T) (U, error)) ([]U, error) {
	us := make([]U, len(ts))

	for i := range ts {
		itm, err := f(ts[i])
		if err != nil {
			return nil, err
		}

		us[i] = itm
	}

	return us, nil
}
