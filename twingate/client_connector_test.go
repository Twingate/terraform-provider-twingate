package twingate

import (
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

// func TestClientConnectorCreateOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Connector Create Ok", func(t *testing.T) {
// 		// response JSON
// 		createConnectorOkJson := `{
// 	  "data": {
// 		"connectorCreate": {
// 		  "entity": {
// 			"id": "test-id",
// 			"name" : "test-name"
// 		  },
// 		  "ok": true,
// 		  "error": null
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createConnectorOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}
// 		remoteNetworkName := "test"

// 		connector, err := client.createConnector(remoteNetworkName)

// 		assert.NoError(t, err)
// 		assert.EqualValues(t, "test-id", connector.ID)
// 		assert.EqualValues(t, "test-name", connector.Name)
// 	})
// }

func TestClientConnectorCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Ok", func(t *testing.T) {
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

		client := newTestClient()
		httpmock.ActivateNonDefault(client.httpClient)
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createConnectorOkJson))
		remoteNetworkName := graphql.String("test")

		connector, err := client.createConnector(remoteNetworkName)

		assert.Nil(t, err)
		assert.EqualValues(t, "test-id", connector.ID)
		assert.EqualValues(t, "test-name", connector.Name)
	})
}

func TestClientConnectorDeleteOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Ok", func(t *testing.T) {
		// response JSON

		// // our database of articles
		deleteConnectorOkJson := `{
		  "data": {
			"connectorDelete": {
			  "ok": true,
			  "error": null
			}
		  }
		}`

		client := newTestClient()
		httpmock.ActivateNonDefault(client.httpClient)
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))

		err := client.deleteConnector(graphql.ID("test"))

		assert.NoError(t, err)
	})
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createNetworkOkJson))
	remoteNetworkName := graphql.String("test")

	remoteNetwork, err := client.createConnector(remoteNetworkName)

	assert.EqualError(t, err, "failed to create connector with id : error_1")
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, deleteConnectorOkJson))
	connectorId := graphql.ID("test")

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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, readNetworkOkJson))
	connectorId := graphql.ID("test")

	connector, err := client.readConnector(connectorId)

	assert.Nil(t, connector)
	assert.EqualError(t, err, "failed to read connector with id test")
}

func TestClientConnectorEmptyDeleteError(t *testing.T) {
	// response JSON
	deleteConnectorOkJson := `{}`

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, deleteConnectorOkJson))
	connectorId := graphql.ID(nil)

	err := client.deleteConnector(connectorId)

	assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "delete", remoteNetworkResourceName, "connectorID").Error())
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

	client := newTestClient()
	httpmock.ActivateNonDefault(client.httpClient)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, readConnectorsOkJson))

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
