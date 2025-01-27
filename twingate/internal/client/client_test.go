package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func newTestClient() *Client {
	return NewClient(
		"twindev.com", "xxxx", "test",
		time.Duration(1)*time.Second, 0, DefaultAgent, "test",
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
		time.Duration(1)*time.Second, 0, DefaultAgent, "test",
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
		time.Duration(10)*time.Second, 3, DefaultAgent, "test",
	)

	internal := client.HTTPClient.Transport.(*retryablehttp.RoundTripper)
	internal.Client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		assert.Less(t, retryNumber, 3)
	}

	_, err := client.post(context.TODO(), "/hello", "hello", nil)

	assert.ErrorContains(t, err, `x509`)
	assert.ErrorContains(t, err, `certificate`)
}

func TestCustomRetryPolicy(t *testing.T) {
	ctx := context.Background()

	// Mock regular expressions used in the function
	certNameNotMatchMacErrorRe = regexp.MustCompile(`certificate name does not match input`)
	certNameNotMatchLinuxErrorRe = regexp.MustCompile(`certificate is valid for`)

	t.Run("No retry on ErrAPITokenNoSet", func(t *testing.T) {
		resp := &http.Response{
			Request: &http.Request{
				Header: http.Header{
					headerRequestID: []string{"test_id"},
				},
			},
		}

		shouldRetry, err := customRetryPolicy(ctx, resp, ErrAPITokenNoSet)

		assert.False(t, shouldRetry)
		assert.Equal(t, ErrAPITokenNoSet, err)
	})

	t.Run("No retry on TLS certificate error", func(t *testing.T) {
		fakeURLError := &url.Error{Err: errors.New("certificate name does not match input")}
		resp := &http.Response{
			Request: &http.Request{
				Header: http.Header{
					headerRequestID: []string{"test_id"},
				},
			},
		}

		shouldRetry, err := customRetryPolicy(ctx, resp, fakeURLError)

		assert.False(t, shouldRetry)
		assert.Equal(t, fakeURLError, err)
	})

	t.Run("Retry enabled on other errors", func(t *testing.T) {
		resp := &http.Response{
			Request: &http.Request{
				Header: http.Header{
					headerRequestID: []string{"test_id"},
				},
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/test/path"},
			},
			Status:     "500 Internal Server Error",
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("test")),
		}

		shouldRetry, err := customRetryPolicy(ctx, resp, errors.New("some network error"))

		assert.True(t, shouldRetry)
		assert.Nil(t, err)
	})

	t.Run("No retry when context is canceled", func(t *testing.T) {
		// Create a canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		resp := &http.Response{
			Request: &http.Request{
				Header: http.Header{
					headerRequestID: []string{"test_id"},
				},
			},
		}

		shouldRetry, err := customRetryPolicy(canceledCtx, resp, nil)

		assert.False(t, shouldRetry)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("Retry logic from DefaultRetryPolicy", func(t *testing.T) {
		resp := &http.Response{
			Request: &http.Request{
				Header: http.Header{
					headerRequestID: []string{"test_id"},
				},
				URL: &url.URL{Path: "/test/path"},
			},
			StatusCode: http.StatusTooManyRequests,
		}

		shouldRetry, err := customRetryPolicy(ctx, resp, nil)

		assert.True(t, shouldRetry)
		assert.Nil(t, err)
	})

	t.Run("Retry request logging", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`{"test": "value"}`)
		resp := &http.Response{
			Request: &http.Request{
				Header: http.Header{
					headerRequestID: []string{"test_id"},
				},
				Body:   io.NopCloser(bytes.NewBuffer(reqBody.Bytes())),
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/test/path"},
			},
			StatusCode: http.StatusGatewayTimeout,
			Body:       io.NopCloser(bytes.NewBufferString("test response body")),
		}

		shouldRetry, err := customRetryPolicy(ctx, resp, errors.New("gateway timeout"))

		assert.True(t, shouldRetry)
		assert.Nil(t, err)
	})
}

func TestCustomRetryPolicy_RequestBodyLogging(t *testing.T) {
	ctx := context.Background()

	requestBody := `{"key":"value"}`
	mockRequest := &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/test/path"},
		Header: http.Header{},
		Body:   io.NopCloser(bytes.NewBufferString(requestBody)),
		GetBody: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBufferString(requestBody)), nil
		},
	}

	mockResponse := &http.Response{
		Request: mockRequest,
	}

	// Capture logs using a bytes buffer to verify log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(io.Discard) // Reset the logger after the test

	_, err := customRetryPolicy(ctx, mockResponse, io.EOF)
	assert.NoError(t, err)

	// Validate logs
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "[WARN] [RETRY_POLICY] [id:test_id] request: "+requestBody)
}
