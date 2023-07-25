//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/component/models/presexch"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/util"
)

func SerializeMatchedSubmissionRequirement(req *presexch.MatchedSubmissionRequirement) (map[string]interface{}, error) {
	descriptors, err := util.MapTo(req.Descriptors, serializeMatchedInputDescriptor)
	if err != nil {
		return nil, fmt.Errorf("descriptors serialization failed: %w", err)
	}

	nested, err := util.MapTo(req.Nested, SerializeMatchedSubmissionRequirement)
	if err != nil {
		return nil, fmt.Errorf("nested requirements serialization failed: %w", err)
	}

	return map[string]interface{}{
		"name":        req.Name,
		"purpose":     req.Purpose,
		"rule":        req.Rule,
		"count":       req.Count,
		"min":         req.Min,
		"max":         req.Max,
		"descriptors": descriptors,
		"nested":      nested,
	}, nil
}

func serializeMatchedInputDescriptor(desc *presexch.MatchedInputDescriptor) (map[string]interface{}, error) {
	vcs, err := util.MapTo(desc.MatchedVCs, SerializeCredential)
	if err != nil {
		return nil, fmt.Errorf("credentials serialization failed: %w", err)
	}

	return map[string]interface{}{
		"id":         desc.ID,
		"name":       desc.Name,
		"purpose":    desc.Purpose,
		"matchedVCs": vcs,
	}, nil
}
