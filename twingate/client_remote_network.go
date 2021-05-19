package twingate

import (
	"fmt"
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

func (client *Client) readAllRemoteNetwork() ([]string, error) { //nolint
	query := map[string]string{
		"query": "{ remoteNetworks { edges { node { id } } } }",
	}

	queryResource, err := client.doGraphqlRequest(query)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	var remoteNetworks = make([]string, 0)

	for _, elem := range queryResource.Path("data.remoteNetworks.edges").Children() {
		nodeID := elem.Path("node.id").Data().(string)
		remoteNetworks = append(remoteNetworks, nodeID)
	}

	return remoteNetworks, nil
}

func (client *Client) readRemoteNetwork(remoteNetworkID string) (*RemoteNetwork, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  remoteNetwork(id: "%s") {
			name
		  }
		}

        `, remoteNetworkID),
	}

	queryRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	remoteNetworkQuery := queryRemoteNetwork.Path("data.remoteNetwork")

	if remoteNetworkQuery.Data() == nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	remoteNetwork := RemoteNetwork{
		ID:   remoteNetworkID,
		Name: remoteNetworkQuery.Path("name").Data().(string),
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
