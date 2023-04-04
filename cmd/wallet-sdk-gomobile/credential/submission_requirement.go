/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
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
	ID         string
	Name       string
	Purpose    string
	MatchedVCs *verifiable.CredentialsArray
}

// Len returns len of wrapper array.
func (s *SubmissionRequirementArray) Len() int {
	return len(s.wrapped)
}

// AtIndex returns item from wrapper array.
func (s *SubmissionRequirementArray) AtIndex(index int) *SubmissionRequirement {
	return &SubmissionRequirement{wrapped: s.wrapped[index]}
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

// DescriptorAtIndex returns submission requirement descriptor at given index.
func (s *SubmissionRequirement) DescriptorAtIndex(index int) *InputDescriptor {
	wrapped := s.wrapped.Descriptors[index]

	descriptor := &InputDescriptor{
		ID:         wrapped.ID,
		Name:       wrapped.Name,
		Purpose:    wrapped.Purpose,
		MatchedVCs: verifiable.NewCredentialsArray(),
	}

	for _, cred := range wrapped.MatchedVCs {
		descriptor.MatchedVCs.Add(verifiable.NewCredential(cred))
	}

	return descriptor
}

// NestedRequirementLength returns submission requirement nested len.
func (s *SubmissionRequirement) NestedRequirementLength() int {
	return len(s.wrapped.Nested)
}

// NestedRequirementAtIndex returns submission requirement nested at given index.
func (s *SubmissionRequirement) NestedRequirementAtIndex(index int) *SubmissionRequirement {
	return &SubmissionRequirement{wrapped: s.wrapped.Nested[index]}
}
