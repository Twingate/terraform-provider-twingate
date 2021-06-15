package twingate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/hasura/go-graphql-client"
)

type RemoteNetwork struct {
	ID   string
	Name string
}

const remoteNetworkResourceName = "remote network"

func (client *Client) createRemoteNetwork(remoteNetworkName string) (*RemoteNetwork, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  remoteNetworkCreate(name: "%s", isActive: true) {
				ok
				error
				entity {
				  id
				}
			  }
		}
        `, remoteNetworkName),
	}

	mutationRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIError(err, "create", remoteNetworkResourceName)
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkCreate.ok").Data().(bool)
	if !status {
		message := mutationRemoteNetwork.Path("data.remoteNetworkCreate.error").Data().(string)

		return nil, NewAPIError(NewMutationError(message), "create", remoteNetworkResourceName)
	}

	remoteNetwork := RemoteNetwork{
		ID: mutationRemoteNetwork.Path("data.remoteNetworkCreate.entity.id").Data().(string),
	}

	return &remoteNetwork, nil
}

func (client *Client) readRemoteNetworks() (map[int]*RemoteNetwork, error) { //nolint
	query := map[string]string{
		"query": "{ remoteNetworks { edges { node { id name } } } }",
	}

	queryResource, err := client.doGraphqlRequest(query)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	var remoteNetworks = make(map[int]*RemoteNetwork)

	queryChildren := queryResource.Path("data.remoteNetworks.edges").Children()

	for i, elem := range queryChildren {
		nodeID := elem.Path("node.id").Data().(string)
		nodeName := elem.Path("node.name").Data().(string)
		c := &RemoteNetwork{ID: nodeID, Name: nodeName}
		remoteNetworks[i] = c
	}

	return remoteNetworks, nil
}

func (client *Client) readRemoteNetwork(remoteNetworkID string) (*RemoteNetwork, error) {
	var q struct {
		RemoteNetwork struct {
			Name graphql.String
		} `graphql:"remoteNetwork(id: $remoteNetworkID)"`
	}

	variables := map[string]interface{}{
		"remoteNetworkID": graphql.ID(remoteNetworkID),
	}

	cl := graphql.NewClient("/query", &http.Client{})

	err := cl.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	j, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("can't parse response body: %w", err)
	}
	parsedResponse, err := gabs.ParseJSON(j)
	if err != nil {
		return nil, fmt.Errorf("can't parse response body: %w", err)
	}

	// queryRemoteNetwork, err := client.doGraphqlRequest(mutation)
	// if err != nil {
	// 	return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	// }

	remoteNetworkQuery := parsedResponse.Path("data.remoteNetwork")

	if remoteNetworkQuery.Data() == nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	remoteNetwork := RemoteNetwork{
		ID:   remoteNetworkID,
		Name: string(q.RemoteNetwork.Name),
	}

	return &remoteNetwork, nil
}

func (client *Client) updateRemoteNetwork(remoteNetworkID, remoteNetworkName string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
				mutation {
					remoteNetworkUpdate(id: "%s", name: "%s"){
						ok
						error
					}
				}
        `, remoteNetworkID, remoteNetworkName),
	}

	mutationRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return NewAPIErrorWithID(err, "update", remoteNetworkResourceName, remoteNetworkID)
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.ok").Data().(bool)
	if !status {
		message := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.error").Data().(string)

		return NewAPIErrorWithID(NewMutationError(message), "update", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}

func (client *Client) deleteRemoteNetwork(remoteNetworkID string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  remoteNetworkDelete(id: "%s"){
			ok
			error
		  }
		}
		`, remoteNetworkID),
	}

	deleteRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	status := deleteRemoteNetwork.Path("data.remoteNetworkDelete.ok").Data().(bool)
	if !status {
		message := deleteRemoteNetwork.Path("data.remoteNetworkDelete.error").Data().(string)

		return NewAPIErrorWithID(NewMutationError(message), "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}
