package client

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/client"
	"github.com/jarcoal/httpmock"
)

func newHTTPMockClient() *client.Client {

	c := client.NewClient("twindev.com", "xxxx", "test",
		time.Duration(1)*time.Second, 2, "test")
	httpmock.ActivateNonDefault(c.HTTPClient)

	return c
}

func MultipleResponders(responses ...httpmock.Responder) httpmock.Responder {
	responseIndex := 0
	mutex := sync.Mutex{}
	return func(req *http.Request) (*http.Response, error) {
		mutex.Lock()
		defer mutex.Unlock()
		defer func() { responseIndex++ }()
		if responseIndex >= len(responses) {
			return nil, fmt.Errorf("not enough responses provided: responder called %d time(s) but %d response(s) provided", responseIndex+1, len(responses))
		}
		res := responses[responseIndex]
		return res(req)
	}
}

func graphqlErr(client *client.Client, message string, err error) string {
	return fmt.Sprintf(`%s: Post "%s": %v`, message, client.GraphqlServerURL, err)
}
