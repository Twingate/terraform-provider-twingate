package twingate

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/stretchr/testify/assert"
)

func TestClientConnectorCreateOk(t *testing.T) {
	// response JSON
	createConnectorOkJson := `{
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

	r := ioutil.NopCloser(bytes.NewReader([]byte(createConnectorOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkName := "test"

	connector, err := client.createConnector(remoteNetworkName)

	assert.NoError(t, err)
	assert.EqualValues(t, "test-id", connector.ID)
	assert.EqualValues(t, "test-name", connector.Name)
}

func TestClientConnectorDeleteOk(t *testing.T) {
	// response JSON
	deleteConnectorOkJson := `{
	  "data": {
		"connectorDelete": {
		  "ok": true,
		  "error": null
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(deleteConnectorOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	connectorId := "test"

	err := client.deleteConnector(connectorId)

	assert.NoError(t, err)
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

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	remoteNetworkName := "test"

	remoteNetwork, err := client.createConnector(remoteNetworkName)

	assert.EqualError(t, err, "failed to create connector: error_1")
	assert.Nil(t, remoteNetwork)
}

func TestClientConnectorDeleteError(t *testing.T) {
	// response JSON
	deleteConnectorOkJson := `{
	  "data": {
		"connectorDelete": {
		  "ok": false,
		  "error": "error_1"
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(deleteConnectorOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	connectorId := "test"

	err := client.deleteConnector(connectorId)

	assert.EqualError(t, err, "failed to delete connector with id test: error_1")
}

func TestClientConnectorReadError(t *testing.T) {
	// response JSON
	readNetworkOkJson := `{
	  "data": {
		"connector": null
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
	connectorId := "id"

	connector, err := client.readConnector(connectorId)

	assert.Nil(t, connector)
	assert.EqualError(t, err, "failed to read connector with id id")
}

func TestClientConnectorReadAllOk(t *testing.T) {
	// response JSON
	readConnectorsOkJson := `{
	  "data": {
		"connectors": {
		  "edges": [
			{
			  "node": {
				"id": "connector1",
				"name": "tf-acc-connector1"
			  }
			},
			{
			  "node": {
				"id": "connector2",
				"name": "connector2"
			  }
			},
			{
			  "node": {
				"id": "connector3",
				"name": "tf-acc-connector3"
			  }
			}
		  ]
		}
	  }
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(readConnectorsOkJson)))
	client := createTestClient()

	GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	connector, err := client.readConnectors()
	assert.NoError(t, err)
	// Resources return dynamic and not ordered object
	// See gabs Children() method.

	r0 := &Connectors{
		ID:   "connector1",
		Name: "tf-acc-connector1",
	}
	r1 := &Connectors{
		ID:   "connector2",
		Name: "connector2",
	}
	r2 := &Connectors{
		ID:   "connector3",
		Name: "tf-acc-connector3",
	}
	mockMap := make(map[int]*Connectors)

	mockMap[0] = r0
	mockMap[1] = r1
	mockMap[2] = r2

	counter := 0
	for _, elem := range connector {
		for _, i := range mockMap {
			if elem.Name == i.Name && elem.ID == i.ID {
				counter++
			}
		}
	}

	if len(mockMap) != counter {
		t.Errorf("Expected map not equal to origin!")
	}
	assert.EqualValues(t, len(mockMap), counter)
}
