package httprequest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trustbloc/logutil-go/pkg/log"
)

var logger = log.New("httprequest")

type Request struct {
	client *http.Client
}

func NewWithHttpClient(client *http.Client) *Request {
	return &Request{client: client}
}

// Send send https request
func (r *Request) Send(method, url, contentType, token string, body io.Reader, responseJSON interface{}) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make http request: %w", err)
	}

	defer closeResponseBody(resp.Body)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, expectedStatusCodeError(http.StatusOK, resp.StatusCode, responseBody)
	}

	if responseJSON != nil {
		if err = json.Unmarshal(responseBody, responseJSON); err != nil {
			return nil, fmt.Errorf("unmarshal initiate oidc4ci resp: %w", err)
		}
	}

	return resp, nil
}

func closeResponseBody(respBody io.Closer) {
	err := respBody.Close()
	if err != nil {
		logger.Error("Failed to close response body", log.WithError(err))
	}
}

func expectedStatusCodeError(expected, actual int, respBytes []byte) error {
	return fmt.Errorf("expected status code %d but got status code %d with response body %s instead",
		expected, actual, respBytes)
}
