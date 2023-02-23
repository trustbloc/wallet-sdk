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

func main() {
	args := os.Args[1:]

	if len(args) >= 2 && args[0] == "issuance" && args[1] != "" {
		initiatePreAuthorizedIssuance(args[1:])
	}

	if len(args) >= 2 && args[0] == "verification" && args[1] != "" {
		initiatePreAuthorizedVerification(args[1:])
	}
}

func initiatePreAuthorizedIssuance(issuerProfileIDs []string) {
	err := testenv.SetupTestEnv("fixtures/keys/tls/ec-cacert.pem")
	if err != nil {
		panic(err)
	}

	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	if err != nil {
		panic(err)
	}

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("test_org")
	if err != nil {
		panic(err)
	}

	var initiateIssuanceURLs []string

	for i := 0; i < len(issuerProfileIDs); i++ {
		for j := 0; j < 120; j++ {
			offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(issuerProfileIDs[i])
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
	err := testenv.SetupTestEnv("fixtures/keys/tls/ec-cacert.pem")
	if err != nil {
		panic(err)
	}

	oidc4vpSetup := oidc4vp.NewSetup(testenv.NewHttpRequest())

	err = oidc4vpSetup.AuthorizeVerifierBypassAuth("test_org")
	if err != nil {
		panic(err)
	}

	var initiateVerificationURLs []string

	for i := 0; i < len(verifierProfileIDs); i++ {
		for j := 0; j < 120; j++ {
			initiateVerificationURL, err := oidc4vpSetup.InitiateInteraction(verifierProfileIDs[i])
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
