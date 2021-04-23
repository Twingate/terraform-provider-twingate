package twingate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	Timeout = 10 * time.Second
)

type HTTPClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

var ErrAPIRequest = errors.New("api request error")

func APIError(format string, a ...interface{}) error {
	a = append([]interface{}{ErrAPIRequest}, a...)
	return fmt.Errorf("%s : "+format, a...)
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
		HTTPClient:   httpClient,
		ServerURL:    serverURL,
		APIServerURL: fmt.Sprintf("%s/api/graphql/", serverURL),
		APIToken:     apiToken,
	}
	log.Printf("[INFO] Using Server URL %s", client.ServerURL)

	return &client
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
	parsedBody, err := client.doGraphqlRequest(jsonData)
	_ = parsedBody
	if err != nil {
		log.Printf("[ERROR] Cannot reach Graphql API Server %s", jsonData)

		return fmt.Errorf("can't parse graphql response: %w", err)
	}
	log.Printf("[INFO] Graphql API Server at URL %s reachable", client.GraphqlServerURL)

	return nil
}
func Check(f func() error) {
	if err := f(); err != nil {
		log.Printf("[ERROR] Error Closing: %s", err)
	}
}

func (client *Client) doRequest(req *retryablehttp.Request) ([]byte, error) {

	req.Header.Set("content-type", "application/json")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't execute http request: %w", err)
	}
	defer Check(res.Body.Close)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response body: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, APIError("request %s failed, status %d, body %s", req.RequestURI, res.StatusCode, body)
	}

	return body, nil
}
func (client *Client) doGraphqlRequest(query map[string]string) (*gabs.Container, error) {
	jsonValue, _ := json.Marshal(query)

	req, err := retryablehttp.NewRequest("POST", client.APIServerURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("could not create GraphQL request : %w", err)

	}

	req.Header.Set("X-API-KEY", client.APIToken)
	body, err := client.doRequest(req)
	_ = body
	if err != nil {
		log.Printf("[ERROR] can't execute request %s", err)

		return nil, fmt.Errorf("can't execute request : %w", err)
	}
	parsedResponse, err := gabs.ParseJSON(body)
	if err != nil {
		log.Printf("[ERROR] Error parsing response %s", string(body))

		return nil, fmt.Errorf("can't parse request body : %w", err)
	}

	if parsedResponse.Path("errors") != nil {
		var messages []string
		for _, child := range parsedResponse.Path("errors").Children() {
			messages = append(messages, child.Path("message").Data().(string))
		}

		return nil, APIError("graphql request returned with errors : %s", strings.Join(messages, ","))
	}

	return parsedResponse, nil
}
