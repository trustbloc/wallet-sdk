/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api_test

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

//go:embed testdata/sample_activity.json
var sampleActivity string

func TestActivity(t *testing.T) {
	t.Run("Using an activity that was created directly in Go", func(t *testing.T) {
		activity := createSampleActivity(t)
		checkActivity(t, activity)

		serializedActivity, err := activity.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedActivity)
	})
	t.Run("Using an activity that was unmarshalled from JSON", func(t *testing.T) {
		activity, err := api.ParseActivity(sampleActivity)
		require.NoError(t, err)
		checkActivity(t, activity)

		serializedActivity, err := activity.Serialize()
		require.NoError(t, err)
		require.NotEmpty(t, serializedActivity)
	})
	t.Run("Fail to unmarshal activity", func(t *testing.T) {
		activity, err := api.ParseActivity("")
		require.EqualError(t, err, "unexpected end of JSON input")
		require.Nil(t, activity)
	})
	t.Run("Unsupported param type", func(t *testing.T) {
		activity := createActivityWithParamsOfUnsupportedTypes()

		valueType, err := activity.Params().GetValueType("unsupportedParam")
		require.EqualError(t, err, "value is of an unsupported type")
		require.Empty(t, valueType)

		valueType, err = activity.Params().GetValueType("someUnsupportedArrayParam")
		require.EqualError(t, err, "value is of an unsupported type")
		require.Empty(t, valueType)
	})
	t.Run("Value is not a string array error when calling GetStringArray on a value that "+
		"is an array of interfaces, but not an array of only strings", func(t *testing.T) {
		activity := createActivityWithParamsOfUnsupportedTypes()

		valueType, err := activity.Params().GetStringArray("someUnsupportedArrayParam")
		require.EqualError(t, err, "value is not an array of strings")
		require.Empty(t, valueType)
	})
}

func createSampleActivity(t *testing.T) *api.Activity {
	t.Helper()

	parsedUUID, err := uuid.Parse("3e983abb-38b4-42be-8634-b4b382f29ee3")
	require.NoError(t, err)

	var unmarshalledTime time.Time

	err = json.Unmarshal([]byte(`"2023-02-03T18:46:44.882178-05:00"`), &unmarshalledTime)
	require.NoError(t, err)

	return &api.Activity{GoAPIActivity: &goapi.Activity{
		ID:   parsedUUID,
		Type: goapi.LogTypeCredentialActivity,
		Time: unmarshalledTime,
		Data: goapi.Data{
			Client:    "https://server.example.com",
			Operation: "oidc-issuance",
			Status:    goapi.ActivityLogStatusSuccess,
			Params: map[string]interface{}{
				"someStringParam":      "someString",
				"someStringArrayParam": []string{"element1", "element2"},
			},
		},
	}}
}

func createActivityWithParamsOfUnsupportedTypes() *api.Activity {
	activity := &api.Activity{GoAPIActivity: &goapi.Activity{
		Data: goapi.Data{
			Params: map[string]interface{}{
				"unsupportedParam":          0,
				"someUnsupportedArrayParam": []interface{}{"someString", 0},
			},
		},
	}}

	return activity
}

func checkActivity(t *testing.T, activity *api.Activity) {
	t.Helper()

	require.Equal(t, "3e983abb-38b4-42be-8634-b4b382f29ee3", activity.ID())
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type())
	require.Equal(t, int64(1675468004), activity.UnixTimestamp())
	require.Equal(t, "https://server.example.com", activity.Client())
	require.Equal(t, "oidc-issuance", activity.Operation())
	require.Equal(t, goapi.ActivityLogStatusSuccess, activity.Status())

	params := activity.Params()

	valueType, err := params.GetValueType("someStringParam")
	require.NoError(t, err)
	require.Equal(t, "string", valueType)

	valueType, err = params.GetValueType("someStringArrayParam")
	require.NoError(t, err)
	require.Equal(t, "[]string", valueType)

	valueType, err = params.GetValueType("nonExistent")
	require.EqualError(t, err, "no value found under the given key")
	require.Empty(t, valueType)

	stringParam, err := params.GetString("someStringParam")
	require.NoError(t, err)
	require.Equal(t, "someString", stringParam)

	stringParam, err = params.GetString("someStringArrayParam")
	require.EqualError(t, err, "value is not a string")
	require.Empty(t, stringParam)

	stringParam, err = params.GetString("nonExistent")
	require.EqualError(t, err, "no value found under the given key")
	require.Empty(t, stringParam)

	stringArrayParam, err := params.GetStringArray("someStringArrayParam")
	require.NoError(t, err)
	checkStringArray(t, stringArrayParam)

	stringArrayParam, err = params.GetStringArray("nonExistent")
	require.EqualError(t, err, "no value found under the given key")
	require.Empty(t, stringArrayParam)

	keyValuePairs := params.AllKeyValuePairs()

	require.Equal(t, 2, keyValuePairs.Length())

	for i := range keyValuePairs.Length() {
		var stringCaseChecked, stringArrayCaseChecked bool

		keyValuePair := keyValuePairs.AtIndex(i)

		if keyValuePair.Key() == "someStringParam" {
			if stringCaseChecked {
				require.FailNow(t, "multiple key-value pairs with the same key")
			}

			valueType, err = keyValuePair.ValueType()
			require.NoError(t, err)
			require.Equal(t, "string", valueType)

			valueAsString, errValueString := keyValuePair.ValueString()
			require.NoError(t, errValueString)
			require.Equal(t, "someString", valueAsString)

			valueAsStringArray, errValueStringArray := keyValuePair.ValueStringArray()
			require.EqualError(t, errValueStringArray, "value is not an array of strings")
			require.Nil(t, valueAsStringArray)

			stringCaseChecked = true
		} else if keyValuePair.Key() == "someStringArrayParam" {
			if stringArrayCaseChecked {
				require.FailNow(t, "multiple key-value pairs with the same key")
			}

			valueType, err = keyValuePair.ValueType()
			require.NoError(t, err)
			require.Equal(t, "[]string", valueType)

			valueAsStringArray, errValueStringArray := keyValuePair.ValueStringArray()
			require.NoError(t, errValueStringArray)
			checkStringArray(t, valueAsStringArray)

			valueAsString, errValueString := keyValuePair.ValueString()
			require.EqualError(t, errValueString, "value is not a string")
			require.Empty(t, valueAsString)

			stringArrayCaseChecked = true
		}
	}

	keyValuePair := keyValuePairs.AtIndex(2)
	require.Nil(t, keyValuePair)

	keyValuePair = keyValuePairs.AtIndex(-1)
	require.Nil(t, keyValuePair)
}

func checkStringArray(t *testing.T, stringArrayParam *api.StringArray) {
	t.Helper()

	require.Equal(t, 2, stringArrayParam.Length())

	require.Equal(t, "element1", stringArrayParam.AtIndex(0))
	require.Equal(t, "element2", stringArrayParam.AtIndex(1))

	require.Equal(t, "", stringArrayParam.AtIndex(2))
	require.Equal(t, "", stringArrayParam.AtIndex(-1))
}
