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
	t.Run("Test Twingate Resource : Client Remote Network Create Ok", func(t *testing.T) {
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

		assert.NoError(t, err)
		assert.EqualValues(t, "test-id", remoteNetwork.ID)
	})
}

func TestClientRemoteNetworkCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Remote Network Create Error", func(t *testing.T) {
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

		assert.EqualError(t, err, "failed to create remote network: error_1")
		assert.Nil(t, remoteNetwork)
	})
}

func TestClientRemoteNetworkUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Remote Network Update Error", func(t *testing.T) {
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

		assert.EqualError(t, err, "failed to update remote network with id id: error_1")
	})
}

func TestClientRemoteNetworkReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Client Remote Network Read Error", func(t *testing.T) {
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
		assert.EqualError(t, err, "failed to read remote network with id id")
	})
}

func TestClientNetworkReadAllOk(t *testing.T) {
	t.Run("Test Twingate Resource : Client Network Read All Ok", func(t *testing.T) {
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

		r := ioutil.NopCloser(bytes.NewReader([]byte(readNetworkOkJson)))
		client := createTestClient()

		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		}

		network, err := client.readRemoteNetworks()
		assert.NoError(t, err)
		// Resources return dynamic and not ordered object
		// See gabs Children() method.

		r0 := &remoteNetwork{
			ID:   "network1",
			Name: "tf-acc-network1",
		}
		r1 := &remoteNetwork{
			ID:   "network2",
			Name: "network2",
		}
		r2 := &remoteNetwork{
			ID:   "network3",
			Name: "tf-acc-network3",
		}
		mockMap := make(map[int]*remoteNetwork)

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
