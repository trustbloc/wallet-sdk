/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialquery allows querying credentials using presentation definition.
package credentialquery

import (
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/bbs-signature-go/bbs12381g2pub"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/proof/defaults"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// Instance implements querying credentials using presentation definition.
type Instance struct {
	documentLoader ld.DocumentLoader
}

type queryOpts struct {
	credentials []*verifiable.Credential

	didResolver              api.DIDResolver
	applySelectiveDisclosure bool
}

// QueryOpt is the query credential option.
type QueryOpt func(opts *queryOpts)

// WithCredentialsArray sets the array of Verifiable Credentials to check against the Presentation Definition.
func WithCredentialsArray(vcs []*verifiable.Credential) QueryOpt {
	return func(opts *queryOpts) {
		opts.credentials = vcs
	}
}

// WithSelectiveDisclosure enables selective disclosure apply.
func WithSelectiveDisclosure(didResolver api.DIDResolver) QueryOpt {
	return func(opts *queryOpts) {
		opts.didResolver = didResolver
		opts.applySelectiveDisclosure = true
	}
}

// NewInstance returns new Instance.
func NewInstance(documentLoader ld.DocumentLoader) *Instance {
	return &Instance{documentLoader: documentLoader}
}

// GetSubmissionRequirements returns information about VCs matching requirements.
func (c *Instance) GetSubmissionRequirements(
	query *presexch.PresentationDefinition,
	opts ...QueryOpt,
) ([]*presexch.MatchedSubmissionRequirement, error) {
	qOpts := &queryOpts{}
	for _, opt := range opts {
		opt(qOpts)
	}

	bbsProofCreator := &verifiable.BBSProofCreator{
		ProofDerivation:            bbs12381g2pub.New(),
		VerificationMethodResolver: common.NewVDRKeyResolver(qOpts.didResolver),
	}

	var matchOpts []presexch.MatchRequirementsOpt
	if qOpts.applySelectiveDisclosure {
		matchOpts = append(matchOpts,
			presexch.WithSelectiveDisclosureApply(),
			presexch.WithSDBBSProofCreator(bbsProofCreator),
			presexch.WithSDCredentialOptions(
				verifiable.WithDisabledProofCheck(),
				verifiable.WithJSONLDDocumentLoader(c.documentLoader),
				verifiable.WithProofChecker(
					defaults.NewDefaultProofChecker(common.NewVDRKeyResolver(qOpts.didResolver))),
			),
		)
	}

	results, err := query.MatchSubmissionRequirement(
		qOpts.credentials,
		c.documentLoader,
		matchOpts...,
	)
	if err != nil {
		return nil,
			walleterror.NewValidationError(
				module,
				FailToGetMatchRequirementsResultsCode,
				FailToGetMatchRequirementsResultsError,
				err)
	}

	return results, nil
}
