package twingate

import (
	"github.com/jarcoal/httpmock"
)

func newHTTPMockClient() *Client {
	sURL := newServerURL("test", "dev.opstg.com")
	client := NewClient(sURL, "xxxx")
	httpmock.ActivateNonDefault(client.RetryableClient.HTTPClient)
	client.RetryableClient.RetryMax = 0

	return client
}
