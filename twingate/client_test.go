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
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
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
	testToken := "token"
	testNetwork := "network"
	testUrl := "twingate.com"

	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	client := NewClient(&testNetwork, &testToken, &testUrl)

	client.HTTPClient = &MockClient{}

	err := client.ping()

	assert.Nil(t, err)
}
