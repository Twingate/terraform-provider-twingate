package twingate

import (
	"fmt"
)

type RemoteNetwork struct {
	Id   string
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
		Id: mutationRemoteNetwork.Path("data.remoteNetworkCreate.entity.id").Data().(string),
	}

	return &remoteNetwork, nil
}

func (client *Client) readRemoteNetwork(remoteNetworkId string) (*RemoteNetwork, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  remoteNetwork(id: "%s") {
			name
		  }
		}

        `, remoteNetworkId),
	}
	queryRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIErrorWithId(err, "read", remoteNetworkResourceName, remoteNetworkId)
	}

	remoteNetworkQuery := queryRemoteNetwork.Path("data.remoteNetwork")

	if remoteNetworkQuery.Data() == nil {
		return nil, NewAPIErrorWithId(err, "read", remoteNetworkResourceName, remoteNetworkId)
	}

	remoteNetwork := RemoteNetwork{
		Id:   remoteNetworkId,
		Name: remoteNetworkQuery.Path("name").Data().(string),
	}

	return &remoteNetwork, nil
}

func (client *Client) updateRemoteNetwork(remoteNetworkId, remoteNetworkName string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
				mutation {
					remoteNetworkUpdate(id: "%s", name: "%s"){
						ok
						error
					}
				}
        `, remoteNetworkId, remoteNetworkName),
	}
	mutationRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return NewAPIErrorWithId(err, "update", remoteNetworkResourceName, remoteNetworkId)
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.ok").Data().(bool)
	if !status {
		message := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.error").Data().(string)

		return NewAPIErrorWithId(NewMutationError(message), "update", remoteNetworkResourceName, remoteNetworkId)
	}

	return nil
}

func (client *Client) deleteRemoteNetwork(remoteNetworkId string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  remoteNetworkDelete(id: "%s"){
			ok
			error
		  }
		}
		`, remoteNetworkId),
	}
	deleteRemoteNetwork, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return NewAPIErrorWithId(err, "delete", remoteNetworkResourceName, remoteNetworkId)
	}

	status := deleteRemoteNetwork.Path("data.remoteNetworkDelete.ok").Data().(bool)
	if !status {
		message := deleteRemoteNetwork.Path("data.remoteNetworkDelete.error").Data().(string)

		return NewAPIErrorWithId(NewMutationError(message), "delete", remoteNetworkResourceName, remoteNetworkId)
	}

	return nil
}
