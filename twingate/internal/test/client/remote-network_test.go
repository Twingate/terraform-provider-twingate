package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientRemoteNetworkCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Ok", func(t *testing.T) {
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, createNetworkOkJson))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), "test")

		assert.Nil(t, err)
		assert.EqualValues(t, "test-id", remoteNetwork.ID)
	})
}

func TestClientRemoteNetworkCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Error", func(t *testing.T) {
		// response JSON
		createNetworkOkJson := `{
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
			httpmock.NewStringResponder(200, createNetworkOkJson))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), "test")

		assert.EqualError(t, err, "failed to create remote network: error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to create remote network: Post "%s": error_1`, client.GraphqlServerURL))
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

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), "test")

		assert.EqualError(t, err, "failed to create remote network: query result is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network Error", func(t *testing.T) {
		// response JSON
		updateNetworkOkJson := `{
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
			httpmock.NewStringResponder(200, updateNetworkOkJson))

		_, err := client.UpdateRemoteNetwork(context.Background(), "id", "test")

		assert.EqualError(t, err, "failed to update remote network with id id: error_1")
	})
}

func TestClientRemoteNetworkUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		_, err := client.UpdateRemoteNetwork(context.Background(), "id", "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to update remote network with id id: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientRemoteNetworkReadByIDError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))

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
			httpmock.NewErrorResponder(errors.New("error_1")))

		remoteNetwork, err := client.ReadRemoteNetworkByID(context.Background(), "id")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read remote network with id id: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientRemoteNetworkReadByNameError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
		    "remoteNetworks": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))

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
			httpmock.NewErrorResponder(errors.New("error_1")))

		remoteNetwork, err := client.ReadRemoteNetworkByName(context.Background(), "name")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read remote network with name name: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientCreateEmptyRemoteNetworkError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Empty Remote Network Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))

		remoteNetwork, err := client.CreateRemoteNetwork(context.Background(), "")

		assert.EqualError(t, err, "failed to create remote network: network name is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientReadEmptyRemoteNetworkByIDError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty Remote Network Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))

		remoteNetwork, err := client.ReadRemoteNetworkByID(context.Background(), "")

		assert.EqualError(t, err, "failed to read remote network: network id is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientReadEmptyRemoteNetworkByNameError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty Remote Network Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
		    "remoteNetworks": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))

		remoteNetwork, err := client.ReadRemoteNetworkByName(context.Background(), "")

		assert.EqualError(t, err, "failed to read remote network: network name is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientDeleteEmptyRemoteNetworkError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Empty Remote Network Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
		    "remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewStringResponder(200, readNetworkOkJson))

		err := client.DeleteRemoteNetwork(context.Background(), "")

		assert.EqualError(t, err, "failed to delete remote network: network id is empty")
	})
}

func TestClientNetworkReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks", func(t *testing.T) {
		expected := []*model.RemoteNetwork{
			{
				ID:   "network1",
				Name: "tf-acc-network1",
			},
			{
				ID:   "network2",
				Name: "network2",
			},
			{
				ID:   "network3",
				Name: "tf-acc-network3",
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
		            "name": "tf-acc-network1"
		          }
		        },
		        {
		          "node": {
		            "id": "network2",
		            "name": "network2"
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
		            "name": "tf-acc-network3"
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

		networks, err := client.ReadRemoteNetworks(context.Background())

		assert.NoError(t, err)
		assert.EqualValues(t, expected, networks)
	})
}

func TestClientNetworkReadAllRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks - Request Error", func(t *testing.T) {
		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			httpmock.NewErrorResponder(errors.New("error_1")))

		networks, err := client.ReadRemoteNetworks(context.Background())

		assert.Nil(t, networks)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read remote network with id All: Post "%s": error_1`, client.GraphqlServerURL))
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

		networks, err := client.ReadRemoteNetworks(context.Background())

		assert.Nil(t, networks)
		assert.EqualError(t, err, `failed to read remote network: query result is empty`)
	})
}

func TestClientNetworkReadAllRequestErrorOnPageFetch(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks - Request Error", func(t *testing.T) {
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

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(200, response1),
				httpmock.NewErrorResponder(errors.New("error_1")),
			),
		)

		networks, err := client.ReadRemoteNetworks(context.Background())

		assert.Nil(t, networks)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read remote network: Post "%s": error_1`, client.GraphqlServerURL))
	})
}
