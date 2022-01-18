package twingate

import (
	"github.com/jarcoal/httpmock"
	"time"
)

func newHTTPMockClient() *Client {

	client := NewClient("twindev.com", "xxxx", "test",
		time.Duration(1) * time.Second, 2, "test")
	httpmock.ActivateNonDefault(client.HTTPClient)

	return client
}


