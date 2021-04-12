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

type Client struct {
	Token      string
	Tenant     string
	ServerURL  string
	HTTPClient *http.Client
}

func NewClient(tenant, token, url *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: TimeOut},
		ServerURL:  fmt.Sprintf("https://%s.%s", *tenant, *url),
		Token:      *token,
	}
	log.Printf("[INFO] Creating Server URL %s", c.ServerURL)
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
	parsedBody, err := c.doGraphqlRequest(jsonData)
	_ = parsedBody
	if err != nil {
		log.Printf("[ERROR] Cannot initialize Grqhql Server %s", jsonData)

		return nil, err
	}

	return &c, nil
}

func Check(f func() error) {
	if err := f(); err != nil {
		log.Printf("[ERROR] Error Closing: %s", err)
	}
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("X-API-KEY", c.Token)
	req.Header.Set("content-type", "application/json")
	res, err := c.HTTPClient.Do(req)
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

func (c *Client) doGraphqlRequest(query map[string]string) (*gabs.Container, error) {
	jsonValue, _ := json.Marshal(query)

	req, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("%s/api/graphql/", c.ServerURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
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
