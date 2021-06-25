package twingate

import (
	"errors"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientRetriesFailedRequestsOnServerError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Retries Failed Requests on Server Error", func(t *testing.T) {
		var serverCallCount int32
		var expectedBody = []byte("Success!")

		testToken := "token"
		testNetwork := "network"
		testUrl := "twingate.com"
		c := http.Client{Transport: newTransport(testToken)}
		sURL := newServerURL(testNetwork, testUrl)
		gql := graphql.NewClient(sURL.newGraphqlServerURL(), &c)
		client := NewClient(sURL, testToken, gql)

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
