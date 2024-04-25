package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/go-uuid"
	"github.com/hasura/go-graphql-client"
)

const (
	EnvPageLimit = "TWINGATE_PAGE_LIMIT"
	EnvAPIToken  = "TWINGATE_API_TOKEN" // #nosec G101

	headerAPIKey        = "X-API-KEY"
	headerAgent         = "User-Agent"
	headerCorrelationID = "X-Correlation-Id"

	defaultPageLimit  = 50
	extendedPageLimit = 100
)

var (
	ErrAPITokenNoSet = errors.New("api_token not set")

	// A regular expression to match the error returned by net/http when the
	// TLS certificate name is not match with input. This error isn't typed
	// specifically so we resort to matching on the error string.
	certNameNotMatchMacErrorRe   = regexp.MustCompile(`certificate name does not match input`)
	certNameNotMatchLinuxErrorRe = regexp.MustCompile(`certificate is valid for`)
)

type Client struct {
	GraphqlClient    *graphql.Client
	HTTPClient       *http.Client
	GraphqlServerURL string
	APIServerURL     string
	version          string
	pageLimit        int
	correlationID    string
}

type transport struct {
	underlineRoundTripper http.RoundTripper
	apiToken              string
	version               string
	correlationID         string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.init(); err != nil {
		return nil, err
	}

	req.Header.Set(headerAPIKey, t.apiToken)
	req.Header.Set(headerAgent, t.version)
	req.Header.Set(headerCorrelationID, t.correlationID)

	return t.underlineRoundTripper.RoundTrip(req) //nolint:wrapcheck
}

func (t *transport) init() error {
	if t.apiToken == "" {
		t.apiToken = os.Getenv(EnvAPIToken)
	}

	if t.apiToken == "" {
		return ErrAPITokenNoSet
	}

	return nil
}

func newTransport(underlineRoundTripper http.RoundTripper, apiToken, version, correlationID string) *transport {
	return &transport{
		underlineRoundTripper: underlineRoundTripper,
		apiToken:              apiToken,
		version:               twingateAgentVersion(version),
		correlationID:         correlationID,
	}
}

func twingateAgentVersion(version string) string {
	return "TwingateTF/" + version
}

func (s *serverURL) newGraphqlServerURL() string {
	return s.url + "/api/graphql/"
}

func (s *serverURL) newAPIServerURL() string {
	return s.url + "/api/v4"
}

type serverURL struct {
	url string
}

func newServerURL(network, url string) serverURL {
	return serverURL{
		url: fmt.Sprintf("https://%s.%s", network, url),
	}
}

func customRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry if API token not set
	if errors.Is(err, ErrAPITokenNoSet) {
		return false, err
	}

	// do not retry if there is an issue with TLS certificate
	if err != nil {
		if v, ok := err.(*url.Error); ok { //nolint:errorlint
			if certNameNotMatchMacErrorRe.MatchString(v.Error()) ||
				certNameNotMatchLinuxErrorRe.MatchString(v.Error()) {
				return false, v
			}
		}
	}

	return retryablehttp.DefaultRetryPolicy(ctx, resp, err) //nolint
}

func NewClient(url string, apiToken string, network string, httpTimeout time.Duration, httpRetryMax int, version string) *Client {
	correlationID, _ := uuid.GenerateUUID()

	sURL := newServerURL(network, url)
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryableClient.CheckRetry = customRetryPolicy
	retryableClient.RetryMax = httpRetryMax
	retryableClient.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		log.Printf("[WARN] Failed to call %s (retry %d)", req.URL.String(), retryNumber)
	}
	retryableClient.HTTPClient.Timeout = httpTimeout
	retryableClient.HTTPClient.Transport = newTransport(retryableClient.HTTPClient.Transport, apiToken, version, correlationID)

	httpClient := retryableClient.StandardClient()

	client := Client{
		HTTPClient:       httpClient,
		GraphqlServerURL: sURL.newGraphqlServerURL(),
		APIServerURL:     sURL.newAPIServerURL(),
		GraphqlClient: graphql.NewClient(sURL.newGraphqlServerURL(), httpClient).WithRequestModifier(func(request *http.Request) {
			request.Header.Set(headerCorrelationID, correlationID)
		}),
		version:       version,
		pageLimit:     getPageLimit(),
		correlationID: correlationID,
	}

	log.Printf("[INFO] Using Server URL %s", sURL.newGraphqlServerURL())

	cache.client = &client

	return &client
}

func getPageLimit() int {
	str := os.Getenv(EnvPageLimit)

	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultPageLimit
	}

	return val
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
	req.Header.Set(headerAgent, twingateAgentVersion(client.version))
	req.Header.Set(headerCorrelationID, client.correlationID)
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

type MutationResponse interface {
	OK() bool
	ErrorStr() string
	ResponseWithPayload
}

func (client *Client) mutate(ctx context.Context, resp MutationResponse, variables map[string]any, opr operation, attrs ...attr) error {
	err := client.GraphqlClient.Mutate(ctx, resp, variables, graphql.OperationName(opr.String()))
	if err != nil {
		return opr.apiError(err, attrs...)
	}

	if !resp.OK() {
		return opr.apiError(NewMutationError(resp.ErrorStr()), attrs...)
	}

	if resp.IsEmpty() {
		return opr.apiError(ErrGraphqlResultIsEmpty, attrs...)
	}

	return nil
}

type ResponseWithPayload interface {
	IsEmpty() bool
}

func (client *Client) query(ctx context.Context, resp ResponseWithPayload, variables map[string]any, opr operation, attrs ...attr) error {
	err := client.GraphqlClient.Query(ctx, resp, variables, graphql.OperationName(opr.String()))
	if err != nil {
		return opr.apiError(err, attrs...)
	}

	if resp.IsEmpty() {
		return opr.apiError(ErrGraphqlResultIsEmpty, attrs...)
	}

	return nil
}
