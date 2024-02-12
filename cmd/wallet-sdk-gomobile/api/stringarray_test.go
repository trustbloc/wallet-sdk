/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

func TestStringArray_Append(t *testing.T) {
	stringArray := api.NewStringArray()

	require.Equal(t, 0, stringArray.Length())

	stringArray.Append("string1")

	require.Equal(t, 1, stringArray.Length())
	require.Equal(t, "string1", stringArray.AtIndex(0))

	stringArray.Append("string2")

	require.Equal(t, 2, stringArray.Length())
	require.Equal(t, "string1", stringArray.AtIndex(0))
	require.Equal(t, "string2", stringArray.AtIndex(1))
}

func TestStringArrayArray_Append(t *testing.T) {
	stringArray := api.NewStringArray()
	stringArrayArray := api.NewStringArrayArray()

	require.Equal(t, 0, stringArray.Length())
	require.Equal(t, 0, stringArrayArray.Length())

	stringArray.Append("string1")
	stringArrayArray.Add(stringArray)

	require.Equal(t, 1, stringArrayArray.Length())
	require.Equal(t, "string1", stringArrayArray.AtIndex(0).AtIndex(0))

	require.Nil(t, stringArrayArray.AtIndex(-1))
}

func TestStringArrayArrayToGoArray(t *testing.T) {
	strings := [][]string{
		{"str11", "str12"},
		{"str21", "str22"},
	}
	result := api.StringArrayArrayToGoArray(api.StringArrayArrayFromGoArray(strings))
	require.ElementsMatch(t, strings, result)
}
