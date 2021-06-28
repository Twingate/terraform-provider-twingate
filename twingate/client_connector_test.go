package twingate

import (
	"errors"
	"net/http"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

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
		deleteConnectorOkJson := `{
		  "data": {
			"connectorDelete": {
			  "ok": true,
			  "error": null
			}
		  }
		}`

		client := newTestClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))

		err := client.deleteConnector(graphql.ID("test"))

		assert.NoError(t, err)
	})
}

func TestClientConnectorCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Error", func(t *testing.T) {

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
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkOkJson))
		remoteNetworkID := graphql.ID("test")

		remoteNetwork, err := client.createConnector(remoteNetworkID)

		assert.EqualError(t, err, "failed to create connector with id : error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Request Error", func(t *testing.T) {

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
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createNetworkOkJson)
				return resp, errors.New("error_1")
			})
		remoteNetworkID := graphql.ID("test")

		remoteNetwork, err := client.createConnector(remoteNetworkID)

		assert.EqualError(t, err, "failed to create connector with id : Post \""+client.GraphqlServerURL+"\": error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorEmptyCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Create Error", func(t *testing.T) {

		// response JSON
		createNetworkOkJson := `{}`

		client := newTestClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkOkJson))
		remoteNetworkID := graphql.ID(nil)

		remoteNetwork, err := client.createConnector(remoteNetworkID)

		assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "create", remoteNetworkResourceName, "remoteNetworkID").Error())
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Error", func(t *testing.T) {

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
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))
		connectorId := graphql.ID("test")

		err := client.deleteConnector(connectorId)

		assert.EqualError(t, err, "failed to delete connector with id test: error_1")
	})
}

func TestClientConnectorDeleteRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Request Error", func(t *testing.T) {

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
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, deleteConnectorOkJson)
				return resp, errors.New("error_1")
			})
		connectorId := graphql.ID("test")

		err := client.deleteConnector(connectorId)

		assert.EqualError(t, err, "failed to delete connector with id test: Post \""+client.GraphqlServerURL+"\": error_1")
	})
}

func TestClientConnectorReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Error", func(t *testing.T) {

		// response JSON
		readNetworkOkJson := `{
	  "data": {
		"connector": null
	  }
	}`

		client := newTestClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))
		connectorId := graphql.ID("test")

		connector, err := client.readConnector(connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to read connector with id test")
	})
}

func TestClientConnectorReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Request Error", func(t *testing.T) {

		// response JSON
		readNetworkOkJson := `{
	  "data": {
		"connector": null
	  }
	}`

		client := newTestClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, readNetworkOkJson)
				return resp, errors.New("error_1")
			})
		connectorId := graphql.ID("test")

		connector, err := client.readConnector(connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to read connector with id test: Post \""+client.GraphqlServerURL+"\": error_1")
	})
}

func TestClientConnectorEmptyReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Read Error", func(t *testing.T) {

		// response JSON
		readNetworkOkJson := `{}`

		client := newTestClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))
		connectorId := graphql.ID(nil)

		connector, err := client.readConnector(connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "read", remoteNetworkResourceName, "connectorID").Error())
	})
}
func TestClientConnectorEmptyDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Delete Error", func(t *testing.T) {

		// response JSON
		deleteConnectorOkJson := `{}`

		client := newTestClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))
		connectorId := graphql.ID(nil)

		err := client.deleteConnector(connectorId)

		assert.EqualError(t, err, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "delete", remoteNetworkResourceName, "connectorID").Error())
	})
}

func TestClientConnectorReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors", func(t *testing.T) {

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
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readConnectorsOkJson))

		connector, err := client.readConnectors()
		assert.NoError(t, err)

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
	})
}
