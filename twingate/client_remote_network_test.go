package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"

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

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkName := "test"

	remoteNetwork, err := client.createRemoteNetwork(remoteNetworkName)

	assert.Nil(t, err)
	assert.EqualValues(t, "test-id", remoteNetwork.Id)
}

func TestClientRemoteNetworkCreateError(t *testing.T) {

	// response JSON
	createNetworkOkJson := `{
	  "data": {
		"remoteNetworkCreate": {
		  "ok": false,
		  "error": "error_1"
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(createNetworkOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkName := "test"

	remoteNetwork, err := client.createRemoteNetwork(remoteNetworkName)

	assert.EqualError(t, err, "api request error : can't create network with name test, error: error_1")
	assert.Nil(t, remoteNetwork)
}

func TestClientRemoteNetworkUpdateError(t *testing.T) {

	// response JSON
	updateNetworkOkJson := `{
	  "data": {
		"remoteNetworkUpdate": {
		  "ok": false,
		  "error": "error_1"
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(updateNetworkOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkId := "id"
	remoteNetworkName := "test-name"

	err := client.updateRemoteNetwork(remoteNetworkId, remoteNetworkName)

	assert.EqualError(t, err, "api request error : can't update network: error_1")
}

func TestClientRemoteNetworkReadError(t *testing.T) {

	// response JSON
	readNetworkOkJson := `{
	  "data": {
		"remoteNetwork": null
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(readNetworkOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkId := "id"

	remoteNetwork, err := client.readRemoteNetwork(remoteNetworkId)

	assert.Nil(t, remoteNetwork)
	assert.EqualError(t, err, "api request error : can't read remote network: id")
}
