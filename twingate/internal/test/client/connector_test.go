package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var (
	notificationEnabled  = true
	notificationDisabled = false
)

func TestClientConnectorCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Ok", func(t *testing.T) {
		expected := &model.Connector{
			ID:                   "test-id",
			Name:                 "test-name",
			NetworkID:            "remote-network-id",
			StatusUpdatesEnabled: &notificationEnabled,
		}

		jsonResponse := `{
		  "data": {
		    "connectorCreate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name",
		        "hasStatusNotificationsEnabled": true,
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
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test"})

		assert.NoError(t, err)
		assert.Equal(t, expected, connector)
	})
}

func TestClientConnectorCreateWithNotificationStatusOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Notification Status Ok", func(t *testing.T) {
		expected := &model.Connector{
			ID:                   "test-id",
			Name:                 "test-name",
			NetworkID:            "remote-network-id",
			StatusUpdatesEnabled: &notificationEnabled,
		}

		jsonResponse := `{
		  "data": {
		    "connectorCreate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name",
		        "hasStatusNotificationsEnabled": true,
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
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.CreateConnector(context.Background(), &model.Connector{
			NetworkID:            "test",
			StatusUpdatesEnabled: &notificationEnabled,
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, connector)
	})
}

func TestClientConnectorCreateWithNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Ok", func(t *testing.T) {
		expected := &model.Connector{
			ID:                   "test-id",
			Name:                 "test-name",
			StatusUpdatesEnabled: &notificationDisabled,
		}

		jsonResponse := `{
		  "data": {
		    "connectorCreate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name",
		        "hasStatusNotificationsEnabled": false
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test", Name: "test-name"})

		assert.NoError(t, err)
		assert.Equal(t, expected, connector)
	})
}

func TestClientConnectorUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Ok", func(t *testing.T) {
		expected := &model.Connector{
			ID:                   "test-id",
			Name:                 "test-name",
			StatusUpdatesEnabled: &notificationEnabled,
		}

		jsonResponse := `{
		  "data": {
		    "connectorUpdate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name",
		        "hasStatusNotificationsEnabled": true
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.UpdateConnector(context.Background(), &model.Connector{ID: "test-id", Name: "test-name"})

		assert.Nil(t, err)
		assert.Equal(t, expected, connector)
	})
}

func TestClientConnectorUpdateWithNotificationStatusOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update With Notification Status Ok", func(t *testing.T) {
		expected := &model.Connector{
			ID:                   "test-id",
			Name:                 "test-name",
			StatusUpdatesEnabled: &notificationEnabled,
		}

		jsonResponse := `{
		  "data": {
		    "connectorUpdate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test-name",
		        "hasStatusNotificationsEnabled": true
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.UpdateConnector(context.Background(), &model.Connector{
			ID: "test-id", Name: "test-name", StatusUpdatesEnabled: &notificationEnabled,
		})

		assert.Nil(t, err)
		assert.Equal(t, expected, connector)
	})
}

func TestClientConnectorDeleteOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Ok", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		err := client.DeleteConnector(context.Background(), "test")

		assert.NoError(t, err)
	})
}

func TestClientConnectorCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test"})

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, "failed to create connector: error_1")
	})
}

func TestClientConnectorCreateWithNotificationStatusError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Notification Status Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.CreateConnector(context.Background(), &model.Connector{
			NetworkID:            "test",
			StatusUpdatesEnabled: &notificationEnabled,
		})

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, "failed to create connector: error_1")
	})
}

func TestClientConnectorCreateWithNameError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test", Name: "test-name"})

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, "failed to create connector with name test-name: error_1")
	})
}

func TestClientConnectorCreateErrorEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Error Empty Result", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test"})

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to create connector: query result is empty")
	})
}

func TestClientConnectorCreateWithNotificationStatusErrorEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Notification Status Error Empty Result", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.CreateConnector(context.Background(), &model.Connector{
			NetworkID:            "test",
			StatusUpdatesEnabled: &notificationEnabled,
		})

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to create connector: query result is empty")
	})
}

func TestClientConnectorCreateWithNameErrorEmptyResult(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Error Empty Result", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test", Name: "test-name"})

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to create connector with name test-name: query result is empty")
	})
}

func TestClientConnectorUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))
		connectorId := "test-id"

		connector, err := client.UpdateConnector(context.Background(), &model.Connector{ID: connectorId, Name: "test-name"})

		assert.Nil(t, connector)
		assert.EqualError(t, err, fmt.Sprintf("failed to update connector with id %s: error_1", connectorId))
	})
}

func TestClientConnectorUpdateErrorWhenIdEmpty(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update Error on empty ID", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		_, err := client.UpdateConnector(context.Background(), nil)

		assert.EqualError(t, err, "failed to update connector: connector id is empty")
	})
}

func TestClientConnectorEmptyNetworkIDCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Network ID Create Error", func(t *testing.T) {
		emptyResponse := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

		connector, err := client.CreateConnector(context.Background(), nil)

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to create connector: network id is empty")
	})
}

func TestClientConnectorDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Error", func(t *testing.T) {
		jsonResponse := `{
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
			httpmock.NewStringResponder(200, jsonResponse))

		connectorId := "test"
		err := client.DeleteConnector(context.Background(), connectorId)

		assert.EqualError(t, err, fmt.Sprintf("failed to delete connector with id %s: error_1", connectorId))
	})
}

func TestClientConnectorReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read - Ok", func(t *testing.T) {
		expected := &model.Connector{
			ID:                   "test-id",
			Name:                 "test-name",
			StatusUpdatesEnabled: &notificationEnabled,
		}

		jsonResponse := `{
		  "data": {
		    "connector": {
		      "id": "test-id",
		      "name": "test-name",
		      "hasStatusNotificationsEnabled": true
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.ReadConnector(context.Background(), "test-id")

		assert.Equal(t, expected, connector)
		assert.NoError(t, err)
	})
}

func TestClientConnectorReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "connector": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connectorId := "test"
		connector, err := client.ReadConnector(context.Background(), connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, fmt.Sprintf("failed to read connector with id %s: query result is empty", connectorId))
	})
}

func TestClientConnectorReadEmptyError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Error", func(t *testing.T) {
		emptyResponse := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

		connectors, err := client.ReadConnectors(context.Background())

		assert.Empty(t, connectors)
		assert.EqualError(t, err, "failed to read connector with id All: query result is empty")
	})
}

func TestClientConnectorEmptyReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Read Error", func(t *testing.T) {
		jsonResponse := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		connector, err := client.ReadConnector(context.Background(), "")

		assert.Nil(t, connector)
		assert.EqualError(t, err, "failed to read connector: id is empty")
	})
}

func TestClientConnectorEmptyDeleteError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Empty Delete Error", func(t *testing.T) {
		emptyResponse := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

		err := client.DeleteConnector(context.Background(), "")

		assert.EqualError(t, err, "failed to delete connector: id is empty")
	})
}

func TestClientConnectorReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors", func(t *testing.T) {
		expected := []*model.Connector{
			{
				ID:                   "connector1",
				Name:                 "tf-acc-connector1",
				StatusUpdatesEnabled: &notificationEnabled,
			},
			{
				ID:                   "connector2",
				Name:                 "connector2",
				StatusUpdatesEnabled: &notificationDisabled,
			},
			{
				ID:                   "connector3",
				Name:                 "tf-acc-connector3",
				StatusUpdatesEnabled: &notificationEnabled,
			},
		}

		jsonResponse := `{
		  "data": {
		    "connectors": {
		      "edges": [
		        {
		          "node": {
		            "id": "connector1",
		            "name": "tf-acc-connector1",
		            "hasStatusNotificationsEnabled": true
		          }
		        },
		        {
		          "node": {
		            "id": "connector2",
		            "name": "connector2",
		            "hasStatusNotificationsEnabled": false
		          }
		        },
		        {
		          "node": {
		            "id": "connector3",
		            "name": "tf-acc-connector3",
		            "hasStatusNotificationsEnabled": true
		          }
		        }
		      ]
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

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
			httpmock.NewErrorResponder(errBadRequest))

		connectorId := "test"
		connector, err := client.UpdateConnector(context.Background(), &model.Connector{ID: connectorId, Name: "new name"})

		assert.Nil(t, connector)
		assert.EqualError(t, err, graphqlErr(client, "failed to update connector with id "+connectorId, errBadRequest))
	})
}

func TestClientConnectorUpdateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Update - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "connectorUpdate": {
		      "entity": null,
		      "ok": true
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))
		connectorId := "test"

		connector, err := client.UpdateConnector(context.Background(), &model.Connector{ID: connectorId, Name: "new name"})

		assert.Nil(t, connector)
		assert.EqualError(t, err, fmt.Sprintf(`failed to update connector with id %s: query result is empty`, connectorId))
	})
}

func TestClientConnectorDeleteRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Delete Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		connectorId := "test"
		err := client.DeleteConnector(context.Background(), connectorId)

		assert.EqualError(t, err, graphqlErr(client, "failed to delete connector with id "+connectorId, errBadRequest))
	})
}

func TestClientConnectorCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		remoteNetwork, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test"})

		assert.EqualError(t, err, graphqlErr(client, "failed to create connector", errBadRequest))
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorCreateWithNotificationStatusRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Notification Status Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		remoteNetwork, err := client.CreateConnector(context.Background(), &model.Connector{
			NetworkID:            "test",
			StatusUpdatesEnabled: &notificationEnabled,
		})

		assert.EqualError(t, err, graphqlErr(client, "failed to create connector", errBadRequest))
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorCreateWithNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Create With Name Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		remoteNetwork, err := client.CreateConnector(context.Background(), &model.Connector{NetworkID: "test", Name: "test-name"})

		assert.EqualError(t, err, graphqlErr(client, "failed to create connector with name test-name", errBadRequest))
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientConnectorReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Connector Read Request Error", func(t *testing.T) {

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))
		connectorId := "test"

		connector, err := client.ReadConnector(context.Background(), connectorId)

		assert.Nil(t, connector)
		assert.EqualError(t, err, graphqlErr(client, "failed to read connector with id "+connectorId, errBadRequest))
	})
}

// readConnectorsWithRemoteNetwork

func TestClientReadConnectorsWithRemoteNetworkOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Client Connectors with remote network - Ok", func(t *testing.T) {
		expected := []*model.Connector{
			{ID: "connector1", Name: "tf-acc-connector1", NetworkID: "tf-acc-network1", StatusUpdatesEnabled: &notificationEnabled},
			{ID: "connector2", Name: "tf-acc-connector2", NetworkID: "tf-acc-network2", StatusUpdatesEnabled: &notificationEnabled},
		}

		jsonResponse := `{
		  "data": {
		    "connectors": {
		      "edges": [
		        {
		          "node": {
		            "id": "connector1",
		            "name": "tf-acc-connector1",
		            "hasStatusNotificationsEnabled": true,
		            "remoteNetwork": {
		              "id": "tf-acc-network1"
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "connector2",
		            "name": "tf-acc-connector2",
		            "hasStatusNotificationsEnabled": true,
		            "remoteNetwork": {
		              "id": "tf-acc-network2"
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
			httpmock.NewStringResponder(200, jsonResponse))

		connectors, err := client.ReadConnectors(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expected, connectors)
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
			httpmock.NewErrorResponder(errBadRequest))

		connectors, err := client.ReadConnectors(context.Background())

		assert.Nil(t, connectors)
		assert.EqualError(t, err, graphqlErr(client, "failed to read connector with id All", errBadRequest))
	})
}

func TestClientReadConnectorsAllPagesOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Pages - Ok", func(t *testing.T) {
		expected := []*model.Connector{
			{ID: "connector1", NetworkID: "tf-acc-network1", Name: "tf-acc-connector1", StatusUpdatesEnabled: &notificationEnabled},
			{ID: "connector2", NetworkID: "tf-acc-network2", Name: "tf-acc-connector2", StatusUpdatesEnabled: &notificationEnabled},
			{ID: "connector3", NetworkID: "tf-acc-network3", Name: "tf-acc-connector3", StatusUpdatesEnabled: &notificationEnabled},
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
		            "hasStatusNotificationsEnabled": true,
		            "remoteNetwork": {
		              "id": "tf-acc-network1"
		            }
		          }
		        },
		        {
		          "node": {
		            "id": "connector2",
		            "name": "tf-acc-connector2",
		            "hasStatusNotificationsEnabled": true,
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
		            "hasStatusNotificationsEnabled": true,
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
				}),
		)

		connectors, err := client.ReadConnectors(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, connectors)
	})
}

func TestClientReadConnectorsAllPagesEmptyResultOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Pages - Empty Result On Fetching ", func(t *testing.T) {
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
		      "edges": []
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewStringResponder(200, nextPage),
			),
		)

		connectors, err := client.ReadConnectors(context.Background())
		assert.Nil(t, connectors)
		assert.EqualError(t, err, `failed to read connector with id All: query result is empty`)
	})
}

func TestClientReadConnectorsAllPagesRequestErrorOnFetching(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Pages - Request Error On Fetching ", func(t *testing.T) {
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
		        }
		      ]
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			),
		)

		connectors, err := client.ReadConnectors(context.Background())
		assert.Nil(t, connectors)
		assert.EqualError(t, err, graphqlErr(client, "failed to read connector with id All", errBadRequest))
	})
}
