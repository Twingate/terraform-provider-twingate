package twingate

import (
	"bytes"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

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

func createTestClient() *Client {

	testToken := "token"
	testNetwork := "network"
	testUrl := "twingate.com"

	mockClient := NewClient(&testNetwork, &testToken, &testUrl)
	mockClient.HTTPClient = &MockClient{}

	return mockClient
}

func TestClientPing(t *testing.T) {

	// response JSON
	json := `{
	  "data": {
		"remoteNetworks": {
		  "edges": [
		  ]
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.Nil(t, err)
}

func TestClientPingRequestFails(t *testing.T) {

	// response JSON
	json := `{}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	GetDoFunc = func(*retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.NotNilf(t, err, "status: 500, body: {} ")
}

func TestClientPingRequestParsingFails(t *testing.T) {

	// response JSON
	json := `{ error }`

	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	GetDoFunc = func(*retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.NotNilf(t, err, "invalid character 'e' looking for beginning of object key string")
}

func TestClientRetriesFailedRequestsOnServerError(t *testing.T) {
	var serverCallCount int32
	var expectedBody = []byte("Success!")

	testToken := "token"
	testNetwork := "network"
	testUrl := "twingate.com"

	client := NewClient(&testNetwork, &testToken, &testUrl)

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
}

func TestClientDoesNotRetryOn400Errors(t *testing.T) {
	var serverCallCount int32
	var expectedBody = []byte("Success!")

	testToken := "token"
	testNetwork := "network"
	testUrl := "twingate.com"

	client := NewClient(&testNetwork, &testToken, &testUrl)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&serverCallCount, 1)
		if atomic.LoadInt32(&serverCallCount) > 1 {
			w.Write(expectedBody)
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(400)
	}))
	defer server.Close()

	req, err := retryablehttp.NewRequest("GET", server.URL+"/some/path", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	body, err := client.doRequest(req)
	if err == nil {
		t.Fatalf("Expected to get an error")
	}

	_ = body
}
