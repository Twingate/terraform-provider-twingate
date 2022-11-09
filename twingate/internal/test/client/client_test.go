package client

import (
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/jarcoal/httpmock"
)

func newHTTPMockClient() *client.Client {

	c := client.NewClient("twindev.com", "xxxx", "test",
		time.Duration(1)*time.Second, 2, "test")
	httpmock.ActivateNonDefault(c.HTTPClient)

	return c
}
