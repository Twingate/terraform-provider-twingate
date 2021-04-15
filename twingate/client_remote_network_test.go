package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientRemoteNetworkCreateOk(t *testing.T) {

	// response JSON
	createNetworkOkJson := `{
	  "data": {
		"remoteNetworkCreate": {
		  "entity": {
			"id": "test-id"
		  },
		  "ok": true,
		  "error": null
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(createNetworkOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkName := "test"

	remoteNetwork, err := client.createRemoteNetwork(&remoteNetworkName)

	assert.Nil(t, err)
	assert.EqualValues(t, "test-id", remoteNetwork.Id)
}

func TestClientRemoteNetworkCreateError(t *testing.T) {

	// response JSON
	createNetworkOkJson := `{
	  "data": {
		"remoteNetworkCreate": {
		  "ok": false,
		  "error": "Cant create network"
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(createNetworkOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkName := "test"

	remoteNetwork, err := client.createRemoteNetwork(&remoteNetworkName)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test")
	assert.Nil(t, remoteNetwork)
}
