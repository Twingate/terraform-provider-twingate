package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

	mockClient := NewClient(testNetwork, testToken, testUrl)
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
	GetDoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.Nil(t, err)
}

func TestInitializeTwingateClientRequestFails(t *testing.T) {

	// response JSON
	json := `{}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.EqualError(t, err, "can't parse graphql response: can't execute request : api request error : request  failed, status 500, body {}")

}

func TestInitializeTwingateClientRequestParsingFails(t *testing.T) {

	// response JSON
	json := `{ error }`

	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	client := createTestClient()

	err := client.ping()

	assert.EqualError(t, err, "can't parse graphql response: can't parse request body : invalid character 'e' looking for beginning of object key string")

}

func TestInitializeTwingateClientGraphqlRequestReturnsErrors(t *testing.T) {

	// response JSON
	json := `{
	  "errors": [
		{
		  "message": "error message",
		  "locations": [
			{
			  "line": 2,
			  "column": 3
			}
		  ],
		  "path": [
			"remoteNetwork"
		  ]
		}
	  ],
	  "data": {
		"remoteNetwork": null
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
	remoteNetworkId := "testId"
	remoteNetwork, err := client.readRemoteNetwork(remoteNetworkId)

	assert.Nil(t, remoteNetwork)
	assert.EqualError(t, err, "can't read remote network : api request error : graphql request returned with errors : error message")

}
