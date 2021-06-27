package twingate

import (
	"errors"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hasura/go-graphql-client"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func newTestClient() *Client {
	sURL := newServerURL("test", "dev.opstg.com")
	client := NewClient(sURL, "xxxx")
	client.HTTPClient = &MockClient{}
	return client
}

// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *retryablehttp.Request) (*http.Response, error)
}

var (
	GetDoFunc func(req *retryablehttp.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *retryablehttp.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func TestClientRetriesFailedRequestsOnServerError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Retries Failed Requests on Server Error", func(t *testing.T) {
		var serverCallCount int32
		var expectedBody = []byte("Success!")

		testToken := "token"
		testNetwork := "network"
		testUrl := "twingate.com"
		sURL := newServerURL(testNetwork, testUrl)
		client := NewClient(sURL, testToken)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&serverCallCount, 1)
			if atomic.LoadInt32(&serverCallCount) > 1 {
				w.Write(expectedBody)
				w.WriteHeader(200)
				return
			}
			w.WriteHeader(500)
		}))
		defer server.Close()

		req, err := retryablehttp.NewRequest("GET", server.URL+"/some/path", nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		body, err := client.doRequest(req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if string(body) != string(expectedBody) {
			t.Fatalf("Wrong body: %v", body)
		}

		if serverCallCount != 2 {
			t.Fatalf("Expected server to be called %d times but it was called %d times", 2, serverCallCount)
		}
	})
}

func TestNewAPIErrorWithID(t *testing.T) {
	t.Run("Test Twingate Resource : New API Error With ID", func(t *testing.T) {
		apiErr := &APIError{
			WrappedError: errors.New("test-error"),
			Operation:    "operation",
			Resource:     "resource",
			ID:           graphql.ID("id"),
		}

		err := NewAPIErrorWithID(errors.New("test-error"), "operation", "resource", graphql.ID("id"))

		assert.Equal(t, apiErr, err)
	})
}

func TestNewAPIError(t *testing.T) {
	t.Run("Test Twingate Resource : New API Error", func(t *testing.T) {
		apiErr := &APIError{
			WrappedError: errors.New("test-error"),
			Operation:    "operation",
			Resource:     "resource",
			ID:           graphql.ID(""),
		}

		err := NewAPIError(errors.New("test-error"), "operation", "resource")

		assert.Equal(t, apiErr.WrappedError, err.WrappedError)
		assert.Equal(t, apiErr.Operation, err.Operation)
		assert.Equal(t, apiErr.Resource, err.Resource)
		assert.Empty(t, err.ID)
	})
}

func TestAPIError(t *testing.T) {
	t.Run("Test Twingate Resource : API Error", func(t *testing.T) {
		apiErr := &APIError{
			WrappedError: errors.New("test-error"),
			Operation:    "operation",
			Resource:     "resource",
			ID:           graphql.ID("id"),
		}

		errString := apiErr.Error()

		assert.Equal(t, "failed to operation resource with id id: test-error", errString)
	})
}

func TestPing(t *testing.T) {
	t.Run("Test Twingate Resource : Ping Error", func(t *testing.T) {
		pingJson := `{}`

		client := newTestClient()
		httpmock.ActivateNonDefault(client.httpClient)
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, pingJson)
				return resp, errors.New("error_1")
			})

		err := client.ping()

		assert.EqualError(t, err, "failed to ping twingate with id : Post \""+client.GraphqlServerURL+"\": error_1")
	})
}
