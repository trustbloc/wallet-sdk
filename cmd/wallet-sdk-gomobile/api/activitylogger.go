/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"encoding/json"
	"errors"
	"fmt"

	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

const (
	noValueFoundErrMsg        = "no value found under the given key"
	unsupportedTypeErrMsg     = "value is of an unsupported type"
	valueNotStringErrMsg      = "value is not a string"
	valueNotStringArrayErrMsg = "value is not an array of strings"
)

// An Activity represents a single activity.
type Activity struct {
	// GoAPIActivity will not be accessible directly in the bindings (will be "skipped").
	// This is OK - this field is not intended to be set directly by the user of the SDK.
	GoAPIActivity *goapi.Activity
}

// ParseActivity parses the given serialized activity and returns an Activity object.
func ParseActivity(activity string) (*Activity, error) {
	var parsedActivity goapi.Activity

	err := json.Unmarshal([]byte(activity), &parsedActivity)
	if err != nil {
		return nil, err
	}

	return &Activity{GoAPIActivity: &parsedActivity}, nil
}

// Serialize serializes this Activity object into JSON.
func (a *Activity) Serialize() (string, error) {
	activityBytes, err := json.Marshal(a.GoAPIActivity)

	return string(activityBytes), err
}

// ID returns this activity's ID.
func (a *Activity) ID() string {
	return a.GoAPIActivity.ID.String()
}

// Type returns this activity's type.
func (a *Activity) Type() string {
	return a.GoAPIActivity.Type
}

// UnixTimestamp returns the time this activity happened as a Unix timestamp.
func (a *Activity) UnixTimestamp() int64 {
	return a.GoAPIActivity.Time.Unix()
}

// Client returns information about with whom this activity was with. For example: the issuer name, verifier name, etc.
func (a *Activity) Client() string {
	return a.GoAPIActivity.Data.Client
}

// Operation returns what operation was performed. For example: oidc-issuance, oidc-presentation, etc.
func (a *Activity) Operation() string {
	return a.GoAPIActivity.Data.Operation
}

// Status returns the status of the operation that was performed. For example: success, failure, etc.
func (a *Activity) Status() string {
	return a.GoAPIActivity.Data.Status
}

// Params returns any additional parameters contained within the activity.
func (a *Activity) Params() *Params {
	return &Params{params: a.GoAPIActivity.Data.Params}
}

// Params represents additional parameters which may be required for wallet applications in the future.
// As such, this is currently a placeholder.
type Params struct {
	params goapi.Params
}

// GetValueType returns the type of the value stored under the given key.
// If there is no param with that key then an error is returned.
// If value is a string, then "string" is returned.
// If value is an array of strings, then "[]string" is returned.
// No other types are (currently) supported - if the value is any other type then an error is returned.
func (p *Params) GetValueType(key string) (string, error) {
	value, exists := p.params[key]
	if !exists {
		return "", errors.New(noValueFoundErrMsg)
	}

	return getType(value)
}

// GetString returns the value stored under the given key.
// If there is no param with that key, or if the param is not a string then an error is returned.
func (p *Params) GetString(key string) (string, error) {
	value, exists := p.params[key]
	if !exists {
		return "", errors.New(noValueFoundErrMsg)
	}

	valueAsString, ok := value.(string)
	if !ok {
		return "", errors.New(valueNotStringErrMsg)
	}

	return valueAsString, nil
}

// GetStringArray returns the value stored under the given key.
// If there is no param with that key, or if the param is not an array of strings then an error is returned.
func (p *Params) GetStringArray(key string) (*StringArray, error) {
	value, exists := p.params[key]
	if !exists {
		return nil, errors.New(noValueFoundErrMsg)
	}

	return interfaceAsStringArray(value)
}

// AllKeyValuePairs returns all params as key-value pairs.
func (p *Params) AllKeyValuePairs() *KeyValuePairs {
	var keyValuePairs []KeyValuePair
	for key, value := range p.params {
		keyValuePairs = append(keyValuePairs, KeyValuePair{
			key:   key,
			value: value,
		})
	}

	return &KeyValuePairs{keyValuePairs: keyValuePairs}
}

// KeyValuePairs represents a set of key-value pairs.
type KeyValuePairs struct {
	keyValuePairs []KeyValuePair
}

// Length returns the number of key-value pairs contained within this KeyValuePairs object.
func (k *KeyValuePairs) Length() int {
	return len(k.keyValuePairs)
}

// AtIndex returns the key-value pair at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (k *KeyValuePairs) AtIndex(index int) *KeyValuePair {
	maxIndex := len(k.keyValuePairs) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &k.keyValuePairs[index]
}

// KeyValuePair represents a single key-value pair.
type KeyValuePair struct {
	key   string
	value interface{}
}

// Key returns this key-value pair's key.
func (k *KeyValuePair) Key() string {
	return k.key
}

// ValueType returns the type of this key-value pair's value.
// If value is a string, then "string" is returned.
// If value is an array of strings, then "[]string" is returned.
// No other types are (currently) supported - if the value is any other type then an error is returned.
func (k *KeyValuePair) ValueType() (string, error) {
	return getType(k.value)
}

// ValueString returns this key-value pair's value.
// If the value is not a string, then an error is returned.
func (k *KeyValuePair) ValueString() (string, error) {
	valueAsString, ok := k.value.(string)
	if !ok {
		return "", errors.New(valueNotStringErrMsg)
	}

	return valueAsString, nil
}

// ValueStringArray returns this key-value pair's value.
// If the value is not an array of strings, then an error is returned.
func (k *KeyValuePair) ValueStringArray() (*StringArray, error) {
	return interfaceAsStringArray(k.value)
}

// getType returns the type of value.
// When unmarshalling an activity from JSON, the Go unmarshaller will use []interface{} even if each element of the
// interface array is a string. This function ensures that "[]string" gets returned regardless of whether value
// is already a []string or is an []interface{} of all strings.
func getType(value interface{}) (string, error) {
	switch typedValue := value.(type) {
	case string, []string:
		return fmt.Sprintf("%T", value), nil
	case []interface{}:
		return getTypeOfInterfaceArray(typedValue)
	default:
		return "", errors.New("value is of an unsupported type")
	}
}

// When unmarshalling an activity from JSON, the Go unmarshaller will use []interface{}.
// This function checks to see if the []interface{} value is really a []string, and if so,
// returns "[]string" (which matches what fmt.Sprintf("%T", value) returns for a []string).
func getTypeOfInterfaceArray(typedValue []interface{}) (string, error) {
	for i := range typedValue {
		_, ok := typedValue[i].(string)
		if !ok {
			return "", errors.New(unsupportedTypeErrMsg)
		}
	}

	return "[]string", nil
}

func interfaceAsStringArray(value interface{}) (*StringArray, error) {
	valueAsStringArray, ok := value.([]string)
	if !ok {
		valueAsInterfaceArray, ok := value.([]interface{})
		if !ok {
			return nil, errors.New(valueNotStringArrayErrMsg)
		}

		strings := make([]string, len(valueAsInterfaceArray))
		for i := range valueAsInterfaceArray {
			strings[i], ok = valueAsInterfaceArray[i].(string)
			if !ok {
				return nil, errors.New(valueNotStringArrayErrMsg)
			}
		}

		return &StringArray{Strings: strings}, nil
	}

	return &StringArray{Strings: valueAsStringArray}, nil
}

// An ActivityLogger logs activities.
type ActivityLogger interface {
	// Log logs a single activity.
	Log(activity *Activity) error
}
