/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

const (
	caCertPath      = "fixtures/keys/tls/ec-cacert.pem"
	vcsAPIDirectURL = "http://localhost:8075"
)

func main() {
	args := os.Args[1:]
	err := testenv.SetupTestEnv(caCertPath)
	if err != nil {
		panic(err)
	}

	if len(args) >= 2 && args[0] == "issuance" && args[1] != "" {
		initiatePreAuthorizedIssuance(args[1:])
	}

	if len(args) >= 2 && args[0] == "verification" && args[1] != "" {
		initiatePreAuthorizedVerification(args[1:])
	}
}

func initiatePreAuthorizedIssuance(issuerProfileIDs []string) {
	driverLicenseClaims := map[string]interface{}{
		"birthdate":            "1990-01-01",
		"document_number":      "123-456-789",
		"driving_privileges":   "G2",
		"expiry_date":          "2025-05-26",
		"family_name":          "Smith",
		"given_name":           "John",
		"issue_date":           "2020-05-27",
		"issuing_authority":    "Ministry of Transport Ontario",
		"issuing_country":      "Canada",
		"resident_address":     "4726 Pine Street",
		"resident_city":        "Toronto",
		"resident_postal_code": "A1B 2C3",
		"resident_province":    "Ontario",
	}

	verifiableEmployeeClaims := map[string]interface{}{
		"displayName":       "John Doe",
		"givenName":         "John",
		"jobTitle":          "Software Developer",
		"surname":           "Doe",
		"preferredLanguage": "English",
		"mail":              "john.doe@foo.bar",
		"photo":             "data-URL-encoded image",
		"sensitiveID":       "123456789",
		"reallySensitiveID": "abcdefg",
	}

	err := testenv.SetupTestEnv("fixtures/keys/tls/ec-cacert.pem")
	if err != nil {
		panic(err)
	}

	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	if err != nil {
		panic(err)
	}

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("test_org", vcsAPIDirectURL)
	if err != nil {
		panic(err)
	}

	var initiateIssuanceURLs []string

	for i := 0; i < len(issuerProfileIDs); i++ {
		for j := 0; j < 120; j++ {
			claims := verifiableEmployeeClaims
			if issuerProfileIDs[i] == "drivers_license_issuer" {
				claims = driverLicenseClaims
			}

			offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(issuerProfileIDs[i], claims)
			if err == nil {
				initiateIssuanceURLs = append(initiateIssuanceURLs, offerCredentialURL)
				break
			}
			println(err.Error())
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		panic(err)
	}

	result := strings.Join(initiateIssuanceURLs, " ")

	fmt.Print(result)
}

func initiatePreAuthorizedVerification(verifierProfileIDs []string) {
	oidc4vpSetup := oidc4vp.NewSetup(testenv.NewHttpRequest())

	err := oidc4vpSetup.AuthorizeVerifierBypassAuth("test_org", vcsAPIDirectURL)
	if err != nil {
		panic(err)
	}

	var initiateVerificationURLs []string

	for i := 0; i < len(verifierProfileIDs); i++ {
		for j := 0; j < 120; j++ {
			initiateVerificationURL, err := oidc4vpSetup.InitiateInteraction(verifierProfileIDs[i], "test purpose.")
			if err == nil {
				initiateVerificationURLs = append(initiateVerificationURLs, initiateVerificationURL)
				break
			}
			println(err.Error())
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		panic(err)
	}

	result := strings.Join(initiateVerificationURLs, " ")

	fmt.Print(result)
}
