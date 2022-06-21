package twingate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

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

		remoteNetwork, err := client.createRemoteNetwork(context.Background(), "test")

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

		remoteNetwork, err := client.createRemoteNetwork(context.Background(), "test")

		assert.EqualError(t, err, "failed to create remote network: error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Remote Network Request Error", func(t *testing.T) {
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
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, createNetworkOkJson)
				return resp, errors.New("error_1")
			})

		remoteNetwork, err := client.createRemoteNetwork(context.Background(), "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to create remote network: Post "%s": error_1`, client.GraphqlServerURL))
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

		err := client.updateRemoteNetwork(context.Background(), "id", "test")

		assert.EqualError(t, err, "failed to update remote network with id id: error_1")
	})
}

func TestClientRemoteNetworkUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update Remote Network Request Error", func(t *testing.T) {
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
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, updateNetworkOkJson)
				return resp, errors.New("error_1")
			})

		err := client.updateRemoteNetwork(context.Background(), "id", "test")

		assert.EqualError(t, err, fmt.Sprintf(`failed to update remote network with id id: Post "%s": error_1`, client.GraphqlServerURL))
	})
}

func TestClientRemoteNetworkReadError(t *testing.T) {
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

		remoteNetwork, err := client.readRemoteNetwork(context.Background(), "id")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, "failed to read remote network with id id: query result is empty")
	})
}

func TestClientRemoteNetworkReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Remote Network Request Error", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
			"remoteNetwork": null
		  }
		}`

		client := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", client.GraphqlServerURL,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, readNetworkOkJson)
				return resp, errors.New("error_1")
			})

		remoteNetwork, err := client.readRemoteNetwork(context.Background(), "id")

		assert.Nil(t, remoteNetwork)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read remote network with id id: Post "%s": error_1`, client.GraphqlServerURL))
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

		remoteNetwork, err := client.createRemoteNetwork(context.Background(), "")

		assert.EqualError(t, err, "failed to create remote network: network name is empty")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientReadEmptyRemoteNetworkError(t *testing.T) {
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

		remoteNetwork, err := client.readRemoteNetwork(context.Background(), "")

		assert.EqualError(t, err, "failed to read remote network: network id is empty")
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

		err := client.deleteRemoteNetwork(context.Background(), "")

		assert.EqualError(t, err, "failed to delete remote network: network id is empty")
	})
}

func TestClientNetworkReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read All Remote Networks", func(t *testing.T) {
		// response JSON
		readNetworkOkJson := `{
		  "data": {
			"remoteNetworks": {
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
				},
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
			httpmock.NewStringResponder(200, readNetworkOkJson))

		network, err := client.readRemoteNetworks(context.Background())
		assert.NoError(t, err)

		r0 := &IDName{
			ID:   "network1",
			Name: "tf-acc-network1",
		}
		r1 := &IDName{
			ID:   "network2",
			Name: "network2",
		}
		r2 := &IDName{
			ID:   "network3",
			Name: "tf-acc-network3",
		}
		mockMap := make(map[int]*IDName)

		mockMap[0] = r0
		mockMap[1] = r1
		mockMap[2] = r2

		counter := 0

		for _, elem := range network {
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
