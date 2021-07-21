package twingate

import (
	"github.com/jarcoal/httpmock"
)

const mockRetries = 0

func newHTTPMockClient() *Client {
	sURL := newServerURL("test", "dev.opstg.com")
	client := NewClient(sURL, "xxxx")
	httpmock.ActivateNonDefault(client.HTTPClient.HTTPClient)
	client.HTTPClient.RetryMax = mockRetries

	return client
}
