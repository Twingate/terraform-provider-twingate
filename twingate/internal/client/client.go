package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/twingate/go-graphql-client"
)

type Client struct {
	GraphqlClient    *graphql.Client
	HTTPClient       *http.Client
	GraphqlServerURL string
	APIServerURL     string
	version          string
}

type transport struct {
	underlineRoundTripper http.RoundTripper
	apiToken              string
	version               string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-API-KEY", t.apiToken)
	req.Header.Set("User-Agent", t.version)

	return t.underlineRoundTripper.RoundTrip(req) //nolint:wrapcheck
}

func newTransport(underlineRoundTripper http.RoundTripper, apiToken string, version string) *transport {
	return &transport{
		underlineRoundTripper: underlineRoundTripper,
		apiToken:              apiToken,
		version:               fmt.Sprintf("TwingateTF/%s", version),
	}
}

func (s *serverURL) newGraphqlServerURL() string {
	return fmt.Sprintf("%s/api/graphql/", s.url)
}

func (s *serverURL) newAPIServerURL() string {
	return fmt.Sprintf("%s/api/v4", s.url)
}

type serverURL struct {
	url string
}

func newServerURL(network, url string) serverURL {
	return serverURL{
		url: fmt.Sprintf("https://%s.%s", network, url),
	}
}

func NewClient(url string, apiToken string, network string, httpTimeout time.Duration, httpRetryMax int, version string) *Client {
	sURL := newServerURL(network, url)
	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryMax = httpRetryMax
	retryableClient.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		log.Printf("[WARN] Failed to call %s (retry %d)", req.URL.String(), retryNumber)
	}
	retryableClient.HTTPClient.Timeout = httpTimeout
	retryableClient.HTTPClient.Transport = newTransport(retryableClient.HTTPClient.Transport, apiToken, version)

	httpClient := retryableClient.StandardClient()

	client := Client{
		HTTPClient:       httpClient,
		GraphqlServerURL: sURL.newGraphqlServerURL(),
		APIServerURL:     sURL.newAPIServerURL(),
		GraphqlClient:    graphql.NewClient(sURL.newGraphqlServerURL(), httpClient),
		version:          version,
	}

	log.Printf("[INFO] Using Server URL %s", sURL.newGraphqlServerURL())

	return &client
}

func (client *Client) post(ctx context.Context, url string, payload interface{}, headers map[string]string) ([]byte, error) {
	var body io.Reader

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.APIServerURL+url,
		body,
	)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	return client.doRequest(req)
}

func (client *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("TwingateTF/%s", client.version))
	res, err := client.HTTPClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("can't execute http request: %w", err)
	}

	defer func(closer io.Closer) {
		if err := closer.Close(); err != nil {
			log.Printf("[ERROR] Error Closing: %s", err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, NewHTTPError(req.URL.String(), res.StatusCode, body)
	}

	return body, nil
}
