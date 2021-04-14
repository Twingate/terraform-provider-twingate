package twingate

import (
	"bytes"
	"context"
	"encoding/json"
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

type Client struct {
	ApiToken     string
	ServerURL    string
	ApiServerURL string
	HTTPClient   HTTPClient
}

func NewClient(network, apiToken, url *string) *Client {
	serverUrl := fmt.Sprintf("https://%s.%s", *network, *url)
	client := Client{
		HTTPClient:   &http.Client{Timeout: TimeOut},
		ServerURL:    serverUrl,
		ApiServerURL: fmt.Sprintf("%s/api/graphql/", serverUrl),
		ApiToken:     *apiToken,
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
	log.Printf("[INFO] Graphql API Server at URL %s reachable", client.ApiServerURL)

	return nil
}
func Check(f func() error) {
	if err := f(); err != nil {
		log.Printf("[ERROR] Error Closing: %s", err)
	}
}

func (client *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("X-API-KEY", client.ApiToken)
	req.Header.Set("content-type", "application/json")
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer Check(res.Body.Close)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s , %w", res.StatusCode, body, err)
	}

	return body, err
}

func (client *Client) doGraphqlRequest(query map[string]string) (*gabs.Container, error) {
	jsonValue, _ := json.Marshal(query)

	req, err := http.NewRequestWithContext(context.Background(), "POST", client.ApiServerURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}

	body, err := client.doRequest(req)
	_ = body
	if err != nil {
		log.Printf("[ERROR] Cant execute request %s", err)

		return nil, err
	}
	parsedResponse, err := gabs.ParseJSON(body)
	if err != nil {
		log.Printf("[ERROR] Error parsing response %s", string(body))

		return nil, err
	}

	return parsedResponse, err
}
