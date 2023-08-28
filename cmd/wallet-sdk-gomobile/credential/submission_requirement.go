/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

// SubmissionRequirement contains information about VCs that matched a presentation definition.
type SubmissionRequirement struct {
	wrapped *presexch.MatchedSubmissionRequirement
}

// SubmissionRequirementArray wrapper around SubmissionRequirement array.
type SubmissionRequirementArray struct {
	wrapped []*presexch.MatchedSubmissionRequirement
}

// InputDescriptor contains information about VCs that matched an input descriptor of presentation definition.
type InputDescriptor struct {
	ID          string
	Name        string
	Purpose     string
	MatchedVCs  *verifiable.CredentialsArray
	constraints *presexch.Constraints
	schemas     []*presexch.Schema
}

// TypeConstraint returns the type constraint specified by this InputDescriptor, if any.
// If there is no type constraint specified, or if it's specified in an unexpected way, then an empty string is
// returned.
func (i *InputDescriptor) TypeConstraint() string {
	if i.constraints != nil {
		for _, field := range i.constraints.Fields {
			typeConstraint, found := getTypeConstraintFromField(field)
			if found {
				return typeConstraint
			}
		}
	}

	return ""
}

// SchemaArray represents an array of Schemas.
type SchemaArray struct {
	schemas []*presexch.Schema
}

// Length returns the number of Schemas contained within this SchemaArray object.
func (s *SchemaArray) Length() int {
	return len(s.schemas)
}

// AtIndex returns the Schema at the given index.
// If the index passed in is out of bounds, or the underlying Schema object at that index is nil, then nil is returned.
func (s *SchemaArray) AtIndex(index int) *Schema {
	maxIndex := len(s.schemas) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	schema := s.schemas[index]

	if schema == nil {
		return nil
	}

	return &Schema{schema: schema}
}

// Schema represents a Schema from an InputDescriptor.
type Schema struct {
	schema *presexch.Schema
}

// URI return's this Schema's URI.
func (s *Schema) URI() string {
	return s.schema.URI
}

// Required returns this Schema's required value.
func (s *Schema) Required() bool {
	return s.schema.Required
}

// Schemas returns the Schemas from this InputDescriptor. If there are none, then nil is returned instead.
func (i *InputDescriptor) Schemas() *SchemaArray {
	if i.schemas == nil {
		return &SchemaArray{}
	}

	return &SchemaArray{schemas: i.schemas}
}

// Len returns len of wrapper array.
func (s *SubmissionRequirementArray) Len() int {
	return len(s.wrapped)
}

// AtIndex returns the SubmissionRequirement at the given index.
// If the index passed in is out of bounds, or the underlying matched submission requirement object at that index is
// nil, then nil is returned.
func (s *SubmissionRequirementArray) AtIndex(index int) *SubmissionRequirement {
	maxIndex := s.Len() - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	matchedSubmissionRequirement := s.wrapped[index]

	if matchedSubmissionRequirement == nil {
		return nil
	}

	return &SubmissionRequirement{wrapped: matchedSubmissionRequirement}
}

// Name returns submission requirement Name.
func (s *SubmissionRequirement) Name() string {
	return s.wrapped.Name
}

// Purpose returns submission requirement Purpose.
func (s *SubmissionRequirement) Purpose() string {
	return s.wrapped.Purpose
}

// Rule returns submission requirement Rule.
func (s *SubmissionRequirement) Rule() string {
	return string(s.wrapped.Rule)
}

// Count returns submission requirement Count.
func (s *SubmissionRequirement) Count() int {
	return s.wrapped.Count
}

// Min returns submission requirement Min.
func (s *SubmissionRequirement) Min() int {
	return s.wrapped.Min
}

// Max returns submission requirement Max.
func (s *SubmissionRequirement) Max() int {
	return s.wrapped.Max
}

// DescriptorLen returns submission requirement descriptor len.
func (s *SubmissionRequirement) DescriptorLen() int {
	return len(s.wrapped.Descriptors)
}

// DescriptorAtIndex returns the submission requirement InputDescriptor at the given index.
// If the index passed in is out of bounds, or the underlying InputDescriptor object at that index is nil,
// then nil is returned.
func (s *SubmissionRequirement) DescriptorAtIndex(index int) *InputDescriptor {
	maxIndex := s.DescriptorLen() - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	inputDescriptor := s.wrapped.Descriptors[index]

	if inputDescriptor == nil {
		return nil
	}

	descriptor := &InputDescriptor{
		ID:          inputDescriptor.ID,
		Name:        inputDescriptor.Name,
		Purpose:     inputDescriptor.Purpose,
		constraints: inputDescriptor.Constraints,
		schemas:     inputDescriptor.Schemas,
		MatchedVCs:  verifiable.NewCredentialsArray(),
	}

	for _, cred := range inputDescriptor.MatchedVCs {
		descriptor.MatchedVCs.Add(verifiable.NewCredential(cred))
	}

	return descriptor
}

// NestedRequirementLength returns submission requirement nested len.
func (s *SubmissionRequirement) NestedRequirementLength() int {
	return len(s.wrapped.Nested)
}

// NestedRequirementAtIndex returns the nested submission requirement at the given index.
// If the index passed in is out of bounds, or the underlying matched submission requirement object at that index is
// nil, then nil is returned.
func (s *SubmissionRequirement) NestedRequirementAtIndex(index int) *SubmissionRequirement {
	maxIndex := s.NestedRequirementLength() - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	matchedSubmissionRequirement := s.wrapped.Nested[index]

	if matchedSubmissionRequirement == nil {
		return nil
	}

	return &SubmissionRequirement{wrapped: matchedSubmissionRequirement}
}

// The bool return value indicates whether a type constraint was found.
func getTypeConstraintFromField(field *presexch.Field) (string, bool) {
	for _, path := range field.Path {
		if path == "$.type" || path == "$.vc.type" {
			constValue, exists := field.Filter.Contains["const"]
			if exists {
				constValueAsString, ok := constValue.(string)
				if ok {
					return constValueAsString, true
				}
			}
		}
	}

	return "", false
}
