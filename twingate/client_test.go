package twingate

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
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

func TestInitializeTwingateClient(t *testing.T) {

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
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.Nil(t, err)
}
