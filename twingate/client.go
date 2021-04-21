package twingate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Jeffail/gabs/v2"
)

const (
	TimeOut = 10 * time.Second
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var ErrAPIRequest = errors.New("api request error")

func APIError(op string) error {
	return fmt.Errorf("APIRequestError %w : %s", ErrAPIRequest, op)
}

type Client struct {
	APIToken     string
	ServerURL    string
	APIServerURL string
	HTTPClient   HTTPClient
}

func NewClient(network, apiToken, url *string) *Client {
	serverURL := fmt.Sprintf("https://%s.%s", *network, *url)
	client := Client{
		HTTPClient:   &http.Client{Timeout: TimeOut},
		ServerURL:    serverURL,
		APIServerURL: fmt.Sprintf("%s/api/graphql/", serverURL),
		APIToken:     *apiToken,
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

		return err
	}
	log.Printf("[INFO] Graphql API Server at URL %s reachable", client.APIServerURL)

	return nil
}
func Check(f func() error) {
	if err := f(); err != nil {
		log.Printf("[ERROR] Error Closing: %s", err)
	}
}

func (client *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("X-API-KEY", client.APIToken)
	req.Header.Set("content-type", "application/json")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cant execute http request : %w", err)
	}
	defer Check(res.Body.Close)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cant read response body : %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, APIError(fmt.Sprintf("request %s failed, status %d, body %s", req.RequestURI, res.StatusCode, body))
	}

	return body, nil
}

func (client *Client) doGraphqlRequest(query map[string]string) (*gabs.Container, error) {
	jsonValue, _ := json.Marshal(query)

	req, err := http.NewRequestWithContext(context.Background(), "POST", client.APIServerURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("could not create request context : %w", err)
	}

	body, err := client.doRequest(req)
	_ = body
	if err != nil {
		log.Printf("[ERROR] Cant execute request %s", err)

		return nil, fmt.Errorf("could not execute request : %w", err)
	}
	parsedResponse, err := gabs.ParseJSON(body)
	if err != nil {
		log.Printf("[ERROR] Error parsing response %s", string(body))

		return nil, fmt.Errorf("could not parse request body : %w", err)
	}

	return parsedResponse, nil
}
