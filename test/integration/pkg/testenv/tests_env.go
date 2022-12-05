package testenv

import (
	"crypto/tls"
	"net/http"

	tlsutils "github.com/trustbloc/cmdutil-go/pkg/utils/tls"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/httprequest"
)

type testEnv struct {
	httpClient *http.Client
}

var testEnvInstance *testEnv

func SetupTestEnv(caCertPath string) error {
	rootCAs, err := tlsutils.GetCertPool(false, []string{caCertPath})
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{RootCAs: rootCAs, MinVersion: tls.VersionTLS12}

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	testEnvInstance = &testEnv{
		httpClient: client,
	}

	return nil
}

func NewHttpRequest() *httprequest.Request {
	return httprequest.NewWithHttpClient(testEnvInstance.httpClient)
}
