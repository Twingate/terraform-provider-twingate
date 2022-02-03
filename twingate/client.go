package twingate

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/twingate/go-graphql-client"
)

type HTTPError struct {
	RequestURI string
	StatusCode int
	Body       []byte
}

func NewHTTPError(requestURI string, statusCode int, body []byte) *HTTPError {
	return &HTTPError{
		RequestURI: requestURI,
		StatusCode: statusCode,
		Body:       body,
	}
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("request %s failed, status %d, body %s", e.RequestURI, e.StatusCode, e.Body)
}

type APIError struct {
	WrappedError error
	Operation    string
	Resource     string
	ID           graphql.ID
}

func NewAPIErrorWithID(wrappedError error, operation string, resource string, id graphql.ID) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
		ID:           id,
	}
}

func NewAPIError(wrappedError error, operation string, resource string) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
		ID:           "",
	}
}

func (e *APIError) Error() string {
	var args = make([]interface{}, 0, 2) //nolint:gomnd
	args = append(args, e.Operation, e.Resource)

	var format = "failed to %s %s"

	if e.ID.(string) != "" {
		format += " with id %s"

		args = append(args, e.ID)
	}

	if e.WrappedError != nil {
		format += ": %s"

		args = append(args, e.WrappedError)
	}

	return fmt.Sprintf(format, args...)
}

type MutationError struct {
	Message graphql.String
}

func NewMutationError(message graphql.String) *MutationError {
	return &MutationError{
		Message: message,
	}
}

func (e *MutationError) Error() string {
	return string(e.Message)
}

type Client struct {
	RetryableClient  *retryablehttp.Client
	GraphqlClient    *graphql.Client
	HTTPClient       *http.Client
	ServerURL        string
	GraphqlServerURL string
	APIServerURL     string
	APIToken         string
	Version          string
}

type transport struct {
	underlineRoundTripper http.RoundTripper
	APIToken              string
	Version               string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-API-KEY", t.APIToken)
	req.Header.Set("User-Agent", t.Version)

	return t.underlineRoundTripper.RoundTrip(req) //nolint:wrapcheck
}

func newTransport(underlineRoundTripper http.RoundTripper, apiToken string, version string) *transport {
	return &transport{
		underlineRoundTripper: underlineRoundTripper,
		APIToken:              apiToken,
		Version:               fmt.Sprintf("TwingateTF/%s", version),
	}
}

func (s *serverURL) newGraphqlServerURL() string {
	return fmt.Sprintf("%s/api/graphql/", s.url)
}

func (s *serverURL) newAPIServerURL() string {
	return fmt.Sprintf("%s/api/v1", s.url)
}

type serverURL struct {
	url string
}

func newServerURL(network, url string) serverURL {
	var s serverURL
	s.url = fmt.Sprintf("https://%s.%s", network, url)

	return s
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
		RetryableClient:  retryableClient,
		HTTPClient:       httpClient,
		ServerURL:        sURL.url,
		GraphqlServerURL: sURL.newGraphqlServerURL(),
		APIServerURL:     sURL.newAPIServerURL(),
		APIToken:         apiToken,
		GraphqlClient:    graphql.NewClient(sURL.newGraphqlServerURL(), httpClient),
		Version:          version,
	}

	log.Printf("[INFO] Using Server URL %s", sURL.newGraphqlServerURL())

	return &client
}

func (client *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("TwingateTF/%s", client.Version))
	res, err := client.HTTPClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("can't execute http request: %w", err)
	}

	defer func(closer io.Closer) {
		if err := closer.Close(); err != nil {
			log.Printf("[ERROR] Error Closing: %s", err)
		}
	}(res.Body)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, NewHTTPError(req.URL.String(), res.StatusCode, body)
	}

	return body, nil
}
