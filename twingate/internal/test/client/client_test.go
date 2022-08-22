package client

import (
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/jarcoal/httpmock"
)

func newHTTPMockClient() *transport.Client {

	client := transport.NewClient("twindev.com", "xxxx", "test",
		time.Duration(1)*time.Second, 2, "test")
	httpmock.ActivateNonDefault(client.HTTPClient)

	return client
}
