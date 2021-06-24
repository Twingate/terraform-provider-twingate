package twingate

import (
	"github.com/hashicorp/go-retryablehttp"

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

		client := NewClient(testNetwork, testToken, testUrl)

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
