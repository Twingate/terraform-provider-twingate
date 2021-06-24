package twingate

import (
	"strconv"
	"testing"
	"time"

	b64 "encoding/base64"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

func TestCreateReadUpdateDeleteOk(t *testing.T) {
	t.Run("Test Twingate : Create Read Update Delete Ok", func(t *testing.T) {

		client, _ := sharedClient("terraformtests")
		ts := time.Now().Unix()
		remoteNetworkName := graphql.String("test-" + strconv.Itoa(int(ts)))

		remoteNetworkCreate, err := client.createRemoteNetwork(remoteNetworkName)

		assert.NoError(t, err)
		assert.NotEmpty(t, remoteNetworkCreate.ID)

		remoteNetworkRead, err := client.readRemoteNetwork(remoteNetworkCreate.ID)

		assert.NoError(t, err)
		assert.EqualValues(t, remoteNetworkCreate.ID, remoteNetworkRead.ID)

		network, err := client.readRemoteNetworks()
		assert.NoError(t, err)
		assert.NotNil(t, network[0])

		connector, err := client.createConnector(remoteNetworkRead.ID)

		assert.NoError(t, err)
		assert.NotEmpty(t, connector.ID)
		assert.NotEmpty(t, connector.Name)

		connectorRead, err := client.readConnector(connector.ID)
		assert.NoError(t, err)
		assert.EqualValues(t, connector.Name, connectorRead.Name)

		connectors, err := client.readConnectors()
		assert.NoError(t, err)
		assert.NotNil(t, connectors[0])

		err = client.generateConnectorTokens(connector)

		assert.NoError(t, err)
		assert.NotEmpty(t, connector.ConnectorTokens.AccessToken)
		assert.NotEmpty(t, connector.ConnectorTokens.RefreshToken)

		err = client.verifyConnectorTokens(connector.ConnectorTokens.RefreshToken, connector.ConnectorTokens.AccessToken)

		assert.NoError(t, err)

		// protocols := newProcolsInput()
		// protocols.TCP.Policy = graphql.String("ALLOW_ALL")
		// protocols.UDP.Policy = graphql.String("ALLOW_ALL")

		// groups := make([]*graphql.ID, 0)
		// group := graphql.ID(b64.StdEncoding.EncodeToString([]byte("testgroup")))
		// groups = append(groups, &group)

		// resourceCreate := &Resource{
		// 	RemoteNetworkID: remoteNetworkRead.ID,
		// 	Address:         graphql.String("test"),
		// 	Name:            graphql.String("testName"),
		// 	GroupsIds:       groups,
		// 	Protocols:       protocols,
		// }

		// err = client.createResource(resourceCreate)

		// assert.NoError(t, err)
		// assert.NotNil(t, resourceCreate.ID)

		// resourceUpdate := &Resource{
		// 	ID:              resourceCreate.ID,
		// 	RemoteNetworkID: remoteNetworkRead.ID,
		// 	Address:         "test.com",
		// 	Name:            "test resource",
		// 	GroupsIds:       resourceCreate.GroupsIds,
		// 	Protocols:       protocols,
		// }

		// resourceRead, err := client.readResource(resourceCreate.ID)

		// assert.NoError(t, err)
		// assert.NotNil(t, resourceRead)

		// err = client.updateResource(resourceUpdate)

		// assert.NoError(t, err)

		// err = client.deleteResource(resourceCreate.ID)

		// assert.NoError(t, err)

		err = client.deleteConnector(connector.ID)

		assert.NoError(t, err)

		err = client.deleteRemoteNetwork(remoteNetworkRead.ID)

		assert.NoError(t, err)

	})
}

func TestCreateReadUpdateDeleteError(t *testing.T) {
	t.Run("Test Twingate : Create Read Update Delete Error", func(t *testing.T) {

		client, _ := sharedClient("terraformtests")
		remoteNetworkName := graphql.String("")

		remoteNetworkCreate, err := client.createRemoteNetwork(remoteNetworkName)

		assert.Error(t, err)
		assert.Nil(t, remoteNetworkCreate)

		remoteNetworkRead, err := client.readRemoteNetwork(graphql.ID("error"))

		assert.Error(t, err)
		assert.Empty(t, remoteNetworkRead)

		protocols := newProcolsInput()
		protocols.TCP.Policy = graphql.String("ALLOW_ALL")
		protocols.UDP.Policy = graphql.String("ALLOW_ALL")

		groups := make([]*graphql.ID, 0)
		group := graphql.ID(b64.StdEncoding.EncodeToString([]byte("testgroup")))
		groups = append(groups, &group)

		resourceCreate := &Resource{
			RemoteNetworkID: graphql.ID("error"),
			Address:         graphql.String("test"),
			Name:            graphql.String("testName"),
			GroupsIds:       groups,
			Protocols:       protocols,
		}

		connectorCreate, err := client.createConnector(graphql.ID("error"))

		assert.Error(t, err)
		assert.Nil(t, connectorCreate)

		connectorRead, err := client.readConnector(graphql.ID(b64.StdEncoding.EncodeToString([]byte("testid1"))))

		assert.Nil(t, connectorRead)
		assert.Error(t, err)

		connector := &Connector{ID: graphql.ID("error")}
		err = client.generateConnectorTokens(connector)

		assert.Error(t, err)
		assert.Nil(t, connector.ConnectorTokens)

		err = client.verifyConnectorTokens("test1", "test2")

		assert.Error(t, err)

		err = client.deleteConnector(graphql.ID("error"))

		assert.Error(t, err)

		err = client.createResource(resourceCreate)

		assert.Error(t, err)
		assert.Nil(t, resourceCreate.ID)

		resourceUpdate := &Resource{
			ID:              graphql.ID("error"),
			RemoteNetworkID: graphql.ID("error"),
			Address:         graphql.String("test.com"),
			Name:            graphql.String("test resource"),
			GroupsIds:       resourceCreate.GroupsIds,
			Protocols:       protocols,
		}

		resource, err := client.readResource(graphql.ID("resource1"))

		assert.Nil(t, resource)
		assert.Error(t, err)

		err = client.updateResource(resourceUpdate)

		assert.Error(t, err)

		err = client.deleteResource(graphql.ID("error"))

		assert.Error(t, err)

		err = client.deleteRemoteNetwork(graphql.ID("error"))

		assert.Error(t, err)

	})
}

// func TestClientRemoteNetworkCreateError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Remote Network Create Error", func(t *testing.T) {
// 		// response JSON
// 		createNetworkOkJson := `{
// 	  "data": {
// 		"remoteNetworkCreate": {
// 		  "ok": false,
// 		  "error": "error_1"
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(createNetworkOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}
// 		remoteNetworkName := "test"

// 		remoteNetwork, err := client.createRemoteNetwork(remoteNetworkName)

// 		assert.EqualError(t, err, "failed to create remote network: error_1")
// 		assert.Nil(t, remoteNetwork)
// 	})
// }

// func TestClientRemoteNetworkUpdateError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Remote Network Update Error", func(t *testing.T) {
// 		// response JSON
// 		updateNetworkOkJson := `{
// 	  "data": {
// 		"remoteNetworkUpdate": {
// 		  "ok": false,
// 		  "error": "error_1"
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(updateNetworkOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}
// 		remoteNetworkId := "id"
// 		remoteNetworkName := "test-name"

// 		err := client.updateRemoteNetwork(remoteNetworkId, remoteNetworkName)

// 		assert.EqualError(t, err, "failed to update remote network with id id: error_1")
// 	})
// }

// func TestClientRemoteNetworkReadError(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Remote Network Read Error", func(t *testing.T) {
// 		// response JSON
// 		readNetworkOkJson := `{
// 	  "data": {
// 		"remoteNetwork": null
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(readNetworkOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}
// 		remoteNetworkId := "id"

// 		remoteNetwork, err := client.readRemoteNetwork(remoteNetworkId)

// 		assert.Nil(t, remoteNetwork)
// 		assert.EqualError(t, err, "failed to read remote network with id id")
// 	})
// }

// func TestClientNetworkReadAllOk(t *testing.T) {
// 	t.Run("Test Twingate Resource : Client Network Read All Ok", func(t *testing.T) {
// 		// response JSON
// 		readNetworkOkJson := `{
// 	  "data": {
// 		"remoteNetworks": {
// 		  "edges": [
// 			{
// 			  "node": {
// 				"id": "network1",
// 				"name": "tf-acc-network1"
// 			  }
// 			},
// 			{
// 			  "node": {
// 				"id": "network2",
// 				"name": "network2"
// 			  }
// 			},
// 			{
// 			  "node": {
// 				"id": "network3",
// 				"name": "tf-acc-network3"
// 			  }
// 			}
// 		  ]
// 		}
// 	  }
// 	}`

// 		r := ioutil.NopCloser(bytes.NewReader([]byte(readNetworkOkJson)))
// 		client := createTestClient()

// 		GetDoFunc = func(req *retryablehttp.Request) (*http.Response, error) {
// 			return &http.Response{
// 				StatusCode: 200,
// 				Body:       r,
// 			}, nil
// 		}

// 		network, err := client.readRemoteNetworks()
// 		assert.NoError(t, err)
// 		// Resources return dynamic and not ordered object
// 		// See gabs Children() method.

// 		r0 := &remoteNetwork{
// 			ID:   "network1",
// 			Name: "tf-acc-network1",
// 		}
// 		r1 := &remoteNetwork{
// 			ID:   "network2",
// 			Name: "network2",
// 		}
// 		r2 := &remoteNetwork{
// 			ID:   "network3",
// 			Name: "tf-acc-network3",
// 		}
// 		mockMap := make(map[int]*remoteNetwork)

// 		mockMap[0] = r0
// 		mockMap[1] = r1
// 		mockMap[2] = r2

// 		counter := 0

// 		for _, elem := range network {
// 			for _, i := range mockMap {
// 				if elem.Name == i.Name && elem.ID == i.ID {
// 					counter++
// 				}
// 			}
// 		}

// 		if len(mockMap) != counter {
// 			t.Errorf("Expected map not equal to origin!")
// 		}
// 		assert.EqualValues(t, len(mockMap), counter)
// 	})
// }
