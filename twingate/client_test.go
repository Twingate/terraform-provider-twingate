package twingate

import (
	"github.com/hashicorp/go-retryablehttp"

	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// // GraphqlMockClient is the mock client
// type MockGraphqlClient struct {
// 	QueryFunc  func(ctx context.Context, m interface{}, variables map[string]interface{}) error
// 	MutateFunc func(ctx context.Context, m interface{}, variables map[string]interface{}) error
// }

// func (gm *MockGraphqlClient) Query(ctx context.Context, m interface{}, variables map[string]interface{}) error {
// 	return GetQueryFunc(ctx, m, variables)
// }

// func (gm *MockGraphqlClient) Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error {
// 	return GetMutateFunc(ctx, m, variables)
// }

// var (
// 	GetQueryFunc  func(ctx context.Context, m interface{}, variables map[string]interface{}) error
// 	GetMutateFunc func(ctx context.Context, m interface{}, variables map[string]interface{}) error
// )

// func createTestGraphqlClient() *Client {

// 	testToken := "token"
// 	testNetwork := "network"
// 	testUrl := "twingate.com"

// 	mockClient := NewClient(testNetwork, testToken, testUrl)

// 	return mockClient
// }

// func TestClientPing(t *testing.T) {

// 	// response JSON
// 	json := `{
// 	  "data": {
// 		"remoteNetworks": {
// 		  "edges": [
// 		  ]
// 		}
// 	  }
// 	}`

// 	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
// 	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 		return &http.Response{
// 			StatusCode: 200,
// 			Body:       r,
// 		}, nil
// 	}
// 	client := createTestClient()

// 	err := client.ping()

// 	assert.Nil(t, err)
// }

// func TestClientPingRequestFails(t *testing.T) {

// 	// response JSON
// 	json := `{}`

// 	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
// 	GetDoFunc = func(*retryablehttp.Request) (*http.Response, error) {
// 		return &http.Response{
// 			StatusCode: 500,
// 			Body:       r,
// 		}, nil
// 	}
// 	client := createTestClient()

// 	err := client.ping()

// 	assert.EqualError(t, err, "failed to ping twingate: can't execute request: request  failed, status 500, body {}")

// }

// func TestClientPingRequestParsingFails(t *testing.T) {

// 	// response JSON
// 	json := `{ error }`

// 	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
// 	GetDoFunc = func(*retryablehttp.Request) (*http.Response, error) {
// 		return &http.Response{
// 			StatusCode: 200,
// 			Body:       r,
// 		}, nil
// 	}
// 	client := createTestClient()

// 	err := client.ping()

// 	assert.EqualError(t, err, "failed to ping twingate: can't parse response body: invalid character 'e' looking for beginning of object key string")

// }

// func TestClientPingRequestParsingFails(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Retries", func(t *testing.T) {

// 		// response JSON
// 		json := `{ error }`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
// 		GetDoFunc = func(*retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}
// 		client := createTestClient()

// 		err := client.ping()

// 		assert.EqualError(t, err, "failed to ping twingate: can't parse response body: invalid character 'e' looking for beginning of object key string")

// 	})
// }

// func TestInitializeTwingateClientGraphqlRequestReturnsErrors(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Initialize Client Request Returns Error", func(t *testing.T) {

// 		// response JSON
// 		json := `{
// 	  "errors": [
// 		{
// 		  "message": "error message",
// 		  "locations": [
// 			{
// 			  "line": 2,
// 			  "column": 3
// 			}
// 		  ],
// 		  "path": [
// 			"remoteNetwork"
// 		  ]
// 		}
// 	  ],
// 	  "data": {
// 		"remoteNetwork": null
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
// 		GetDoFunc = func(*retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}
// 		client := createTestClient()
// 		remoteNetworkId := "testId"
// 		remoteNetwork, err := client.readRemoteNetwork(remoteNetworkId)

// 		assert.Nil(t, remoteNetwork)
// 		assert.EqualError(t, err, "failed to read remote network with id testId: graphql errors: error message")
// 	})
// }

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

// func TestClientDoesNotRetryOn400Errors(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Doesn't Retry on 400 Errors", func(t *testing.T) {
// 		var serverCallCount int32
// 		var expectedBody = []byte("Success!")

// 		testToken := "token"
// 		testNetwork := "network"
// 		testUrl := "twingate.com"

// 		client := NewClient(testNetwork, testToken, testUrl)

// 		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			atomic.AddInt32(&serverCallCount, 1)
// 			if atomic.LoadInt32(&serverCallCount) > 1 {
// 				w.Write(expectedBody)
// 				w.WriteHeader(200)
// 				return
// 			}
// 			w.WriteHeader(400)
// 		}))
// 		defer server.Close()

// 		req, err := retryablehttp.NewRequest("GET", server.URL+"/some/path", nil)
// 		if err != nil {
// 			t.Fatalf("err: %v", err)
// 		}

// 		body, err := client.doRequest(req)
// 		if err == nil {
// 			t.Fatalf("Expected to get an error")
// 		}

// 		_ = body
// 	})
// }
