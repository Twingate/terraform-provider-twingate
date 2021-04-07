package twingate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	Token      string
	Tenant     string
	ServerURL  string
	HTTPClient *http.Client
}

func NewClient(tenant, token, url *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		ServerURL:  fmt.Sprintf("https:%s.%s", *tenant, *url),
	}

	if (token != nil) && (tenant != nil) {
		jsonData := map[string]string{
			"query": `
            { 
                remoteNetworks {
                    edges {
						id
					}
                }
            }
        `,
		}
		jsonValue, _ := json.Marshal(jsonData)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/graphql/", c.ServerURL), bytes.NewBuffer(jsonValue))

		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req)
		_ = body
		if err != nil {
			return nil, err
		}

	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("X-API-KEY", c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
