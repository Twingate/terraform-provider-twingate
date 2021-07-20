package twingate

import (
	"github.com/jarcoal/httpmock"
)

const mockRetries = 0

func newHTTPMockClient() *Client {
	client := NewClient("dev.opstg.com", "xxxx", "test")
	httpmock.ActivateNonDefault(client.HTTPClient.HTTPClient)
	client.HTTPClient.RetryMax = mockRetries

	return client
}
