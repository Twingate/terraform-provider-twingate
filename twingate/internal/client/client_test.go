package client

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func newTestClient() *Client {
	return NewClient(
		"twindev.com", "xxxx", "test",
		time.Duration(1)*time.Second, 0, "test",
	)
}

func TestNewClientPayloadMarshalError(t *testing.T) {
	c := newTestClient()
	_, err := c.post(context.TODO(), "/hello", make(chan int), nil)

	assert.ErrorContains(t, err, "json")
}

func TestNewClientCancelledContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	c := newTestClient()
	_, err := c.post(ctx, "/hello", "hello", nil)

	assert.ErrorContains(t, err, "can't execute http request")
}

func TestNewClientNilContextError(t *testing.T) {
	c := newTestClient()
	_, err := c.post(nil, "/hello", "hello", nil)

	assert.ErrorContains(t, err, "net/http: nil Context")
}

type dummyFailingReadCloser struct{}

func (d *dummyFailingReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("bad error")
}

func (d *dummyFailingReadCloser) Close() error {
	return nil
}

func TestClientFailedReadBody(t *testing.T) {
	client := newTestClient()
	httpmock.ActivateNonDefault(client.HTTPClient)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", client.APIServerURL+"/hello",
		func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Body: &dummyFailingReadCloser{},
			}, nil
		})

	_, err := client.post(context.TODO(), "/hello", "hello", nil)

	assert.ErrorContains(t, err, "can't read response body")
}

func TestClientAPITokenNotSet(t *testing.T) {
	apiToken := os.Getenv(EnvAPIToken)
	os.Setenv(EnvAPIToken, "")
	defer os.Setenv(EnvAPIToken, apiToken)

	client := NewClient(
		"twindev.com", "", "test",
		time.Duration(1)*time.Second, 0, "test",
	)

	_, err := client.post(context.TODO(), "/hello", "hello", nil)

	assert.ErrorContains(t, err, ErrAPITokenNoSet.Error())

	os.Setenv(EnvAPIToken, "xxx")
	_, err = client.post(context.TODO(), "/hello", "hello", nil)

	assert.ErrorContains(t, err, "lookup test.twindev.com")
}

func TestClientInvalidServerAddress(t *testing.T) {
	client := NewClient(
		"beamreach.twingate.com", "XXXXX", "beamreach",
		time.Duration(10)*time.Second, 3, "test",
	)

	internal := client.HTTPClient.Transport.(*retryablehttp.RoundTripper)
	internal.Client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		assert.Less(t, retryNumber, 3)
	}

	_, err := client.post(context.TODO(), "/hello", "hello", nil)

	assert.ErrorContains(t, err, `x509`)
	assert.ErrorContains(t, err, `certificate`)
}
