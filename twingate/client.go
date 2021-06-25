package twingate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hasura/go-graphql-client"
)

const (
	Timeout = 10 * time.Second
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
	var a = make([]interface{}, 0, 2) //nolint:gomnd
	a = append(a, e.Operation, e.Resource)

	var format = "failed to %s %s"

	if e.ID != 0 || e.ID != nil {
		format += " with id %s"

		a = append(a, e.ID)
	}

	if e.WrappedError != nil {
		format += ": %s"

		a = append(a, e.WrappedError)
	}

	return fmt.Sprintf(format, a...)
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

type HTTPClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

type Client struct {
	GraphqlClient    Gql
	HTTPClient       HTTPClient
	ServerURL        string
	GraphqlServerURL string
	APIServerURL     string
	APIToken         string
	httpClient       *http.Client
}

type transport struct {
	underlyingTransport http.RoundTripper
	APIToken            string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-API-KEY", t.APIToken)

	return t.underlyingTransport.RoundTrip(req) //nolint:wrapcheck
}

func newTransport(apiToken string) *transport {
	return &transport{
		underlyingTransport: http.DefaultTransport,
		APIToken:            apiToken,
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

//go:generate mockgen -destination=../mock/graphql.go -package=mock_twingate github.com/Twingate/terraform-provider-twingate Gql

type Gql interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	NamedQuery(ctx context.Context, name string, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error
	NamedMutate(ctx context.Context, name string, m interface{}, variables map[string]interface{}) error
	QueryRaw(ctx context.Context, q interface{}, variables map[string]interface{}) (*json.RawMessage, error)
	NamedQueryRaw(ctx context.Context, name string, q interface{}, variables map[string]interface{}) (*json.RawMessage, error)
	MutateRaw(ctx context.Context, m interface{}, variables map[string]interface{}) (*json.RawMessage, error)
	NamedMutateRaw(ctx context.Context, name string, m interface{}, variables map[string]interface{}) (*json.RawMessage, error)
}

func NewClient(sURL serverURL, apiToken string, gql Gql) *Client {
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Timeout = Timeout
	httpClient.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		log.Printf("[WARN] Failed to call %s (retry %d)", req.URL.String(), retryNumber)
	}

	client := Client{
		HTTPClient:       httpClient,
		ServerURL:        sURL.url,
		GraphqlServerURL: sURL.newGraphqlServerURL(),
		APIServerURL:     sURL.newAPIServerURL(),
		APIToken:         apiToken,
		GraphqlClient:    graphql.NewClient(graphqlServerURL, &c),
		httpClient:       &c,
	}

	log.Printf("[INFO] Using Server URL %s", sURL.newGraphqlServerURL())

	return &client
}

type pingQuery struct {
	RemoteNetworks struct {
		Edges []Edges
	}
}

func (client *Client) ping() error {
	r := pingQuery{}
	variables := map[string]interface{}{}

	err := client.GraphqlClient.Query(context.Background(), &r, variables)
	if err != nil {
		log.Printf("[ERROR] Cannot reach Graphql API Server %s", client.APIServerURL)

		return NewAPIError(err, "ping", "twingate")
	}

	log.Printf("[INFO] Graphql API Server at URL %s reachable", client.GraphqlServerURL)

	return nil
}

func (client *Client) doRequest(req *retryablehttp.Request) ([]byte, error) {
	req.Header.Set("content-type", "application/json")
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
		return nil, NewHTTPError(req.RequestURI, res.StatusCode, body)
	}

	return body, nil
}
