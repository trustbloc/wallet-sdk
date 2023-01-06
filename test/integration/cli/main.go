package main

import (
	"fmt"
	"os"
	"time"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

func main() {
	args := os.Args[1:]

	if len(args) == 1 && args[0] == "issuance" {
		initiatePreAuthorizedIssuance()
	}

	if len(args) == 1 && args[0] == "verification" {
		initiatePreAuthorizedVerification()
	}
}

func initiatePreAuthorizedIssuance() {
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

	initiateIssuanceURL := ""

	for i := 0; i < 120; i++ {
		initiateIssuanceURL, err = oidc4ciSetup.InitiatePreAuthorizedIssuance("bank_issuer")
		if err == nil {
			break
		}
		println(err.Error())
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		panic(err)
	}

	fmt.Print(initiateIssuanceURL)
}

func initiatePreAuthorizedVerification() {
	err := testenv.SetupTestEnv("fixtures/keys/tls/ec-cacert.pem")
	if err != nil {
		panic(err)
	}

	oidc4vpSetup := oidc4vp.NewSetup(testenv.NewHttpRequest())

	err = oidc4vpSetup.AuthorizeVerifierBypassAuth("test_org")
	if err != nil {
		panic(err)
	}

	initiateIssuanceURL := ""

	for i := 0; i < 120; i++ {
		initiateIssuanceURL, err = oidc4vpSetup.InitiateInteraction("v_myprofile_jwt")
		if err == nil {
			break
		}
		println(err.Error())
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		panic(err)
	}

	fmt.Print(initiateIssuanceURL)
}
