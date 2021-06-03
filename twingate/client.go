package twingate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	Timeout = 10 * time.Second
)

type HTTPClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

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

type GraphQLError struct {
	Messages []string
}

func NewGraphQLError(messages []string) *GraphQLError {
	return &GraphQLError{
		Messages: messages,
	}
}

func (e *GraphQLError) Error() string {
	return fmt.Sprintf("graphql errors: %s", strings.Join(e.Messages, ","))
}

type APIError struct {
	WrappedError error
	Operation    string
	Resource     string
	ID           string
}

func NewAPIErrorWithID(wrappedError error, operation string, resource string, id string) *APIError {
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
	var a = make([]interface{}, 0, 2)
	a = append(a, e.Operation, e.Resource)

	var format = "failed to %s %s"

	if len(e.ID) > 0 {
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
	Message string
}

func NewMutationError(message string) *MutationError {
	return &MutationError{
		Message: message,
	}
}

func (e *MutationError) Error() string {
	return e.Message
}

type Client struct {
	APIToken         string
	ServerURL        string
	GraphqlServerURL string
	APIServerURL     string
	HTTPClient       HTTPClient
}

func NewClient(network, apiToken, url string) *Client {
	serverURL := fmt.Sprintf("https://%s.%s", network, url)

	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Timeout = Timeout
	httpClient.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		log.Printf("[WARN] Failed to call %s (retry %d)", req.URL.String(), retryNumber)
	}

	client := Client{
		HTTPClient:       httpClient,
		ServerURL:        serverURL,
		GraphqlServerURL: fmt.Sprintf("%s/api/graphql/", serverURL),
		APIServerURL:     fmt.Sprintf("%s/api/v1", serverURL),
		APIToken:         apiToken,
	}
	log.Printf("[INFO] Using Server URL %s", client.ServerURL)

	return &client
}

type pingResponse struct {
	Errors string `json:"error"`
	Data   struct {
		Remotenetworks struct {
			Edges []*struct {
				Name string `json:"name"`
			} `json:"edges"`
		} `json:"remoteNetworks"`
	} `json:"data"`
}

func (client *Client) ping() error {
	jsonData := map[string]string{
		"query": `
			{
			  remoteNetworks {
				edges {
				  node {
					id
				  }
				}
			  }
			}
        `,
	}

	r := pingResponse{}
	err := client.doGraphqlRequest(jsonData, &r)

	if err != nil {
		log.Printf("[ERROR] Cannot reach Graphql API Server %s", jsonData)

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

func (client *Client) doGraphqlRequest(query map[string]string, v interface{}) error {
	jsonValue, _ := json.Marshal(query)

	req, err := retryablehttp.NewRequest("POST", client.GraphqlServerURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("could not create GraphQL request : %w", err)
	}

	req.Header.Set("X-API-KEY", client.APIToken)

	body, err := client.doRequest(req)
	if err != nil {
		return fmt.Errorf("can't execute request: %w", err)
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		log.Println("body " + string(body))
		return fmt.Errorf("can't parse response body: %w", err)
	}

	return nil
}
