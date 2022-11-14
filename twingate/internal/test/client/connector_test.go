package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
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
		        "name": "test-name",
		        "remoteNetwork": {
		          "id": "remote-network-id"
		        }
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createConnectorOkJson))

		connector, err := client.CreateConnector(context.Background(), "test", "")

		assert.Nil(t, err)
		assert.EqualValues(t, "test-id", connector.ID)
		assert.EqualValues(t, "test-name", connector.Name)
	})
}

func TestClientConnectorCreateWithNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Ok", func(t *testing.T) {
		// response JSON
		createConnectorOkJson := `{
		  "data": {
		    "connectorCreate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createConnectorOkJson))

		connector, err := client.CreateConnector(context.Background(), "test", "test-name")

		assert.Nil(t, err)
		assert.EqualValues(t, "test-id", connector.ID)
		assert.EqualValues(t, "test-name", connector.Name)
	})
}

func TestClientConnectorUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Ok", func(t *testing.T) {
		// response JSON
		updateConnectorOkJson := `{
		  "data": {
		    "connectorUpdate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, updateConnectorOkJson))

		_, err := client.UpdateConnector(context.Background(), "test-id", "test-name")

		assert.Nil(t, err)
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))

		err := client.DeleteConnector(context.Background(), "test")

		assert.NoError(t, err)
	})
}

func TestClientConnectorCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Error", func(t *testing.T) {

		// response JSON
		createNetworkErrorJson := `{
		  "data": {
		    "connectorCreate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkErrorJson))

		remoteNetwork, err := client.CreateConnector(context.Background(), "test", "")

		assert.EqualError(t, err, "failed to create connector: error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorCreateWithNameError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Error", func(t *testing.T) {

		// response JSON
		createNetworkErrorJson := `{
		  "data": {
		    "connectorCreate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkErrorJson))

		remoteNetwork, err := client.CreateConnector(context.Background(), "test", "test-name")

		assert.EqualError(t, err, "failed to create connector with name test-name: error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorCreateErrorEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Error Empty Result", func(t *testing.T) {

		// response JSON
		createNetworkErrorJson := `{
		  "data": {
		    "connectorCreate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkErrorJson))

		connector, err := client.CreateConnector(context.Background(), "test", "")

		assert.EqualError(t, err, "failed to create connector: query result is empty")
		assert.Nil(t, connector)
	})
}

func TestClientConnectorCreateWithNameErrorEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Error Empty Result", func(t *testing.T) {

		// response JSON
		createNetworkErrorJson := `{
		  "data": {
		    "connectorCreate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkErrorJson))

		connector, err := client.CreateConnector(context.Background(), "test", "test-name")

		assert.EqualError(t, err, "failed to create connector with name test-name: query result is empty")
		assert.Nil(t, connector)
	})
}

func TestClientConnectorUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Error", func(t *testing.T) {

		// response JSON
		createNetworkOkJson := `{
		  "data": {
		    "connectorUpdate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkOkJson))
		connectorId := "test-id"

		_, err := client.UpdateConnector(context.Background(), connectorId, "test-name")

		assert.EqualError(t, err, fmt.Sprintf("failed to update connector with id %s: error_1", connectorId))
	})
}

func TestClientConnectorUpdateErrorWhenIdEmpty(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Error on empty ID", func(t *testing.T) {

		// response JSON
		createNetworkOkJson := `{
		  "data": {
		    "connectorUpdate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkOkJson))

		_, err := client.UpdateConnector(context.Background(), "", "")

		assert.EqualError(t, err, "failed to update connector: connector id is empty")
	})
}

func TestClientConnectorEmptyNetworkIDCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Network ID Create Error", func(t *testing.T) {

		// response JSON
		createNetworkOkJson := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkOkJson))

		remoteNetwork, err := client.CreateConnector(context.Background(), "", "")

		assert.EqualError(t, err, "failed to create connector: network id is empty")
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))
		connectorId := "test"

		err := client.DeleteConnector(context.Background(), connectorId)

		assert.EqualError(t, err, fmt.Sprintf("failed to delete connector with id %s: error_1", connectorId))
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))
		connectorId := "test"

		connector, err := client.ReadConnector(context.Background(), connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, fmt.Sprintf("failed to read connector with id %s: query result is empty", connectorId))
	})
}

func TestClientConnectorReadEmptyError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Error", func(t *testing.T) {

		// response JSON
		readConnectorFAIL := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readConnectorFAIL))

		connectors, _ := client.ReadConnectors(context.Background())

		assert.Empty(t, connectors)
	})
}

func TestClientConnectorEmptyReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Read Error", func(t *testing.T) {

		// response JSON
		readConnectorOkJson := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readConnectorOkJson))

		connector, err := client.ReadConnector(context.Background(), "")

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to read connector: id is empty")
	})
}

func TestClientConnectorEmptyDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Delete Error", func(t *testing.T) {

		// response JSON
		deleteConnectorOkJson := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, deleteConnectorOkJson))

		err := client.DeleteConnector(context.Background(), "")

		assert.EqualError(t, err, "failed to delete connector: id is empty")
	})
}

func TestClientConnectorReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors", func(t *testing.T) {
		expected := []*model.Connector{
			{
				ID:   "connector1",
				Name: "tf-acc-connector1",
			},
			{
				ID:   "connector2",
				Name: "connector2",
			},
			{
				ID:   "connector3",
				Name: "tf-acc-connector3",
			},
		}

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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readConnectorsOkJson))

		connectors, err := client.ReadConnectors(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, connectors)
	})
}

func TestClientConnectorUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))
		connectorId := "test"

		_, err := client.UpdateConnector(context.Background(), connectorId, "new name")

		assert.EqualError(t, err, fmt.Sprintf(`failed to update connector with id %s: Post "%s": error_1`, connectorId, client.GraphqlServerURL))
	})
}

func TestClientConnectorDeleteRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))
		connectorId := "test"

		err := client.DeleteConnector(context.Background(), connectorId)

		assert.EqualError(t, err, fmt.Sprintf(`failed to delete connector with id %s: Post "%s": error_1`, connectorId, client.GraphqlServerURL))
	})
}

func TestClientConnectorCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		remoteNetwork, err := client.CreateConnector(context.Background(), "test", "")

		assert.EqualError(t, err, fmt.Sprintf(`failed to create connector: Post "%s": error_1`, client.GraphqlServerURL))
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorCreateWithNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		remoteNetwork, err := client.CreateConnector(context.Background(), "test", "test-name")

		assert.EqualError(t, err, fmt.Sprintf(`failed to create connector with name test-name: Post "%s": error_1`, client.GraphqlServerURL))
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))
		connectorId := "test"

		connector, err := client.ReadConnector(context.Background(), connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read connector with id %s: Post "%s": error_1`, connectorId, client.GraphqlServerURL))
	})
}

// readConnectorsWithRemoteNetwork

func TestClientReadConnectorsWithRemoteNetworkOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors with remote network - Ok", func(t *testing.T) {
		data := []struct {
			id        string
			name      string
			networkID string
		}{
			{id: "connector1", name: "tf-acc-connector1", networkID: "tf-acc-network1"},
			{id: "connector2", name: "tf-acc-connector2", networkID: "tf-acc-network2"},
		}

		jsonResponse := fmt.Sprintf(`{
		  "data": {
		    "connectors": {
		      "edges": [
		        {
		          "node": {
		            "id": "%s",
		            "name": "%s",
		            "remoteNetwork": {
		              "id": "%s"
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "%s",
		            "name": "%s",
		            "remoteNetwork": {
		              "id": "%s"
		            }
		          }
		        }
		      ]
		    }
		  }
		}`, data[0].id, data[0].name, data[0].networkID, data[1].id, data[1].name, data[1].networkID)

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connectors, err := client.ReadConnectors(context.Background())
		assert.NoError(t, err)

		for i, elem := range connectors {
			assert.EqualValues(t, data[i].id, elem.ID)
			assert.EqualValues(t, data[i].name, elem.Name)
			assert.EqualValues(t, data[i].networkID, elem.NetworkID)
		}
	})
}

func TestClientReadConnectorsWithRemoteNetworkError(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors with remote network - Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "connectors": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connectors, err := client.ReadConnectors(context.Background())

		assert.Nil(t, connectors)
		assert.EqualError(t, err, "failed to read connector with id All: query result is empty")
	})
}

func TestClientReadConnectorsWithRemoteNetworkRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors with remote network - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		connectors, err := client.ReadConnectors(context.Background())

		assert.Nil(t, connectors)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read connector with id All: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientReadConnectorsAllPagesOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Pages - Ok", func(t *testing.T) {
		expected := []*model.Connector{
			{ID: "connector1", NetworkID: "tf-acc-network1", Name: "tf-acc-connector1"},
			{ID: "connector2", NetworkID: "tf-acc-network2", Name: "tf-acc-connector2"},
			{ID: "connector3", NetworkID: "tf-acc-network3", Name: "tf-acc-connector3"},
		}

		jsonResponse := `{
		  "data": {
		    "connectors": {
		      "pageInfo": {
		        "endCursor": "cursor001",
		        "hasNextPage": true
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "connector1",
		            "name": "tf-acc-connector1",
		            "remoteNetwork": {
		              "id": "tf-acc-network1"
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "connector2",
		            "name": "tf-acc-connector2",
		            "remoteNetwork": {
		              "id": "tf-acc-network2"
		            }
		          }
		        }
		      ]
		    }
		  }
		}`

		nextPage := `{
		  "data": {
		    "connectors": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "connector3",
		            "name": "tf-acc-connector3",
		            "remoteNetwork": {
		              "id": "tf-acc-network3"
		            }
		          }
		        }
		      ]
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.ResponderFromMultipleResponses(
				[]*http.Response{
					httpmock.NewStringResponse(200, jsonResponse),
					httpmock.NewStringResponse(200, nextPage),
				},
				t.Log),
		)

		connectors, err := client.ReadConnectors(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, connectors)
	})
}
