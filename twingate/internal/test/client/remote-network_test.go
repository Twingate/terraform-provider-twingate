package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientRemoteNetworkCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Ok", func(t *testing.T) {
		expected := &model.RemoteNetwork{
			ID:       "test-id",
			Name:     "test",
			Location: model.LocationOther,
		}

		jsonResponse := `{
		  "data": {
		    "remoteNetworkCreate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test",
		        "location": "OTHER"
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

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, remoteNetwork)
	})
}

func TestClientRemoteNetworkCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkCreate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.EqualError(t, err, "failed to create remote network with name test: error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.EqualError(t, err, graphqlErr(client, "failed to create remote network with name test", errBadRequest))
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkCreateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkCreate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.EqualError(t, err, "failed to create remote network with name test: query result is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkUpdate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		_, err := client.UpdateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			ID:       "id",
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.EqualError(t, err, "failed to update remote network with id id: error_1")
	})
}

func TestClientRemoteNetworkUpdateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkUpdate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		_, err := client.UpdateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			ID:       "id",
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.EqualError(t, err, "failed to update remote network with id id: query result is empty")
	})
}

func TestClientRemoteNetworkUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network - Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkUpdate": {
		      "ok": true,
		      "entity": {
		        "id": "network-id",
		        "name": "test",
		        "location": "OTHER"
		      }
		    }
		  }
		}`

		expected := &model.RemoteNetwork{
			ID:       "network-id",
			Name:     "test",
			Location: model.LocationOther,
		}

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		network, err := client.UpdateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			ID:       "network-id",
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, network)
	})
}

func TestClientRemoteNetworkUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		_, err := client.UpdateRemoteNetwork(context.Background(), &model.RemoteNetwork{
			ID:       "id",
			Name:     "test",
			Location: model.LocationOther,
		})

		assert.EqualError(t, err, graphqlErr(client, "failed to update remote network with id id", errBadRequest))
	})
}

func TestClientRemoteNetworkReadByIDError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.ReadRemoteNetworkByID(context.Background(), "id")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, "failed to read remote network with id id: query result is empty")
	})
}

func TestClientRemoteNetworkReadByIDRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		remoteNetwork, err := client.ReadRemoteNetworkByID(context.Background(), "id")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, graphqlErr(client, "failed to read remote network with id id", errBadRequest))
	})
}

func TestClientRemoteNetworkReadByNameError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworks": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.ReadRemoteNetworkByName(context.Background(), "name")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, "failed to read remote network with name name: query result is empty")
	})
}

func TestClientRemoteNetworkReadByNameRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		remoteNetwork, err := client.ReadRemoteNetworkByName(context.Background(), "name")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, graphqlErr(client, "failed to read remote network with name name", errBadRequest))
	})
}

func TestClientCreateEmptyRemoteNetworkError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Empty Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), &model.RemoteNetwork{})

		assert.EqualError(t, err, "failed to create remote network: network name is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientReadEmptyRemoteNetworkByIDError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.ReadRemoteNetworkByID(context.Background(), "")

		assert.EqualError(t, err, "failed to read remote network: network id is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientReadEmptyRemoteNetworkByNameError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty Remote Network Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworks": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		remoteNetwork, err := client.ReadRemoteNetworkByName(context.Background(), "")

		assert.EqualError(t, err, "failed to read remote network: network name is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientDeleteRemoteNetworkWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Remote Network - With Empty ID", func(t *testing.T) {
		jsonResponse := `{}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := client.DeleteRemoteNetwork(context.Background(), "")

		assert.EqualError(t, err, "failed to delete remote network: network id is empty")
	})
}

func TestClientDeleteRemoteNetworkRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Remote Network - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := client.DeleteRemoteNetwork(context.Background(), "network-id")

		assert.EqualError(t, err, graphqlErr(client, "failed to delete remote network with id network-id", errBadRequest))
	})
}

func TestClientDeleteRemoteNetworkError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Remote Network - Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkDelete": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := client.DeleteRemoteNetwork(context.Background(), "network-id")

		assert.EqualError(t, err, "failed to delete remote network with id network-id: error_1")
	})
}

func TestClientDeleteRemoteNetworkOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Remote Network - Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworkDelete": {
		      "ok": true
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := client.DeleteRemoteNetwork(context.Background(), "network-id")

		assert.NoError(t, err)
	})
}

func TestClientNetworkReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks", func(t *testing.T) {
		expected := []*model.RemoteNetwork{
			{
				ID:       "network1",
				Name:     "tf-acc-network1",
				Location: model.LocationAzure,
			},
			{
				ID:       "network2",
				Name:     "network2",
				Location: model.LocationAWS,
			},
			{
				ID:       "network3",
				Name:     "tf-acc-network3",
				Location: model.LocationGoogleCloud,
			},
		}

		response1 := `{
		  "data": {
		    "remoteNetworks": {
		      "pageInfo": {
		        "hasNextPage": true,
		        "endCursor": "cur-001"
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "network1",
		            "name": "tf-acc-network1",
		            "location": "AZURE"
		          }
		        },
		        {
		          "node": {
		            "id": "network2",
		            "name": "network2",
		            "location": "AWS"
		          }
		        }
		      ]
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "remoteNetworks": {
		      "pageInfo": {
		        "hasNextPage": false
		      },
		      "edges": [
		        {
		          "node": {
		            "id": "network3",
		            "name": "tf-acc-network3",
		            "location": "GOOGLE_CLOUD"
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
				httpmock.NewStringResponder(200, response1),
				httpmock.NewStringResponder(200, response2),
			),
		)

		networks, err := client.ReadRemoteNetworks(context.Background(), "", "")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, networks)
	})
}

func TestClientNetworkReadAllRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		networks, err := client.ReadRemoteNetworks(context.Background(), "", "")

		assert.Nil(t, networks)
		assert.EqualError(t, err, graphqlErr(client, "failed to read remote network with id All", errBadRequest))
	})
}

func TestClientNetworkReadAllEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks - Empty Response", func(t *testing.T) {
		response1 := `{
		  "data": {
		    "remoteNetworks": {
		      "pageInfo": {
		        "hasNextPage": true,
		        "endCursor": "cur-001"
		      },
		      "edges": [{}]
		    }
		  }
		}`

		response2 := `{
		  "data": {
		    "remoteNetworks": {
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
				httpmock.NewStringResponder(200, response1),
				httpmock.NewStringResponder(200, response2),
			),
		)

		networks, err := client.ReadRemoteNetworks(context.Background(), "", "")

		assert.Nil(t, networks)
		assert.EqualError(t, err, `failed to read remote network: query result is empty`)
	})
}

func TestClientNetworkReadAllRequestErrorOnPageFetch(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks - Request Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "remoteNetworks": {
		      "pageInfo": {
		        "hasNextPage": true,
		        "endCursor": "cur-001"
		      },
		      "edges": [{}]
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

		networks, err := client.ReadRemoteNetworks(context.Background(), "", "")

		assert.Nil(t, networks)
		assert.EqualError(t, err, graphqlErr(client, "failed to read remote network", errBadRequest))
	})
}

func TestClientReadRemoteNetworkWithIDOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network With ID - Ok", func(t *testing.T) {
		expected := &model.RemoteNetwork{
			ID:       "network1",
			Name:     "tf-acc-network1",
			Location: model.LocationOther,
		}

		jsonResponse := `{
		  "data": {
		    "remoteNetwork": {
		      "id": "network1",
		      "name": "tf-acc-network1",
		      "location": "OTHER"
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse),
		)

		network, err := client.ReadRemoteNetwork(context.Background(), "network1", "")

		assert.NoError(t, err)
		assert.Equal(t, expected, network)
	})
}

func TestClientReadRemoteNetworkWithNameOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network With Name - Ok", func(t *testing.T) {
		expected := &model.RemoteNetwork{
			ID:       "network1",
			Name:     "tf-acc-network1",
			Location: model.LocationAWS,
		}

		jsonResponse := `{
		  "data": {
		    "remoteNetworks": {
		      "edges": [
		        {
		          "node": {
		            "id": "network1",
		            "name": "tf-acc-network1",
		            "location": "AWS"
		          }
		        },
		        {
		          "node": {
		            "id": "network2",
		            "name": "tf-acc-network1",
		            "location": "AZURE"
		          }
		        }
		      ]
		    }
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse),
		)

		network, err := client.ReadRemoteNetwork(context.Background(), "", "tf-acc-network1")

		assert.NoError(t, err)
		assert.Equal(t, expected, network)
	})
}
