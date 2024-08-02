/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/trustbloc/logutil-go/pkg/log"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/logfields"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

const (
	composeDir      = "./fixtures/"
	composeFilePath = composeDir + "docker-compose.yml"
	caCertPath      = "fixtures/keys/tls/ec-cacert.pem"
	vcsAPIDirectURL = "http://localhost:8075"
	didResolverURL  = "http://did-resolver.trustbloc.local:8072/1.0/identifiers"
	attestationURL  = "https://localhost:8097/profiles/profileID/profileVersion/wallet/attestation/"
)

var logger = log.New("wallet-sdk-integration-test")

func TestMain(m *testing.M) {
	// Pre setup
	InitializeTestSuite()
	code := m.Run()
	Teardown()
	os.Exit(code)
}

func InitializeTestSuite() {
	err := testenv.SetupTestEnv(caCertPath)
	if err != nil {
		logger.Fatal("setup test environment", log.WithError(err))
	}

	if os.Getenv("ENABLE_COMPOSITION") == "true" {
		runDockerCompose()
	}
}

func Teardown() {
	if os.Getenv("ENABLE_COMPOSITION") == "true" {
		stopDockerCompose()
	}
}

func runDockerCompose() {
	dockerComposeUp := []string{"docker", "compose", "-f", composeFilePath, "up", "--force-recreate", "-d"}

	logger.Info("Running ", logfields.WithDockerComposeCmd(strings.Join(dockerComposeUp, " ")))

	cmd := exec.Command(dockerComposeUp[0], dockerComposeUp[1:]...) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Fatal("test runDockerCompose", logfields.WithCommand(string(out)), log.WithError(err))
	}

	testSleep := 60

	if os.Getenv("TEST_SLEEP") != "" {
		s, err := strconv.Atoi(os.Getenv("TEST_SLEEP"))
		if err != nil {
			logger.Error("invalid 'TEST_SLEEP'", log.WithError(err))
		} else {
			testSleep = s
		}
	}

	sleepD := time.Second * time.Duration(testSleep)
	logger.Info("*** testSleep", logfields.WithSleep(sleepD))
	time.Sleep(sleepD)
}

func stopDockerCompose() {
	if os.Getenv("DISABLE_COMPOSE") == "true" {
		return
	}

	dockerComposeDown := []string{"docker", "compose", "-f", composeFilePath, "down"}

	logger.Info("Running ", logfields.WithDockerComposeCmd(strings.Join(dockerComposeDown, " ")))

	cmd := exec.Command(dockerComposeDown[0], dockerComposeDown[1:]...) //nolint:gosec
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Fatal("test stopDockerCompose", logfields.WithCommand(string(out)), log.WithError(err))
	}
}
