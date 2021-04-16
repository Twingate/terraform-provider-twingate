package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientConnectorCreateOk(t *testing.T) {

	// response JSON
	createNetworkOkJson := `{
	  "data": {
		"connectorCreate": {
		  "entity": {
			"id": "test-id",
			"name" : "test-name"
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

	connector, err := client.createConnector(&remoteNetworkName)

	assert.Nil(t, err)
	assert.EqualValues(t, "test-id", connector.Id)
	assert.EqualValues(t, "test-name", connector.Name)
}

func TestClientConnectorCreateError(t *testing.T) {

	// response JSON
	createNetworkOkJson := `{
	  "data": {
		"connectorCreate": {
		  "ok": false,
		  "error": "error_1"
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

	remoteNetwork, err := client.createConnector(&remoteNetworkName)

	assert.EqualError(t, err, "Cant create connector under the network with name test, Error:  error_1")
	assert.Nil(t, remoteNetwork)
}
