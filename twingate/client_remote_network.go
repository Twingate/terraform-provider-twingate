package twingate

import (
	"fmt"
)

type RemoteNetwork struct {
	Id   string
	Name string
}

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
		return nil, fmt.Errorf("can't create network : %w", err)
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkCreate.ok").Data().(bool)
	if !status {
		errorMessage := mutationRemoteNetwork.Path("data.remoteNetworkCreate.error").Data().(string)

		return nil, APIError("can't create network with name %s, error: %s", remoteNetworkName, errorMessage)
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
		return nil, fmt.Errorf("can't read remote network : %w", err)
	}

	remoteNetworkQuery := queryRemoteNetwork.Path("data.remoteNetwork")

	if remoteNetworkQuery.Data() == nil {
		return nil, APIError("can't read remote network: %s", remoteNetworkId)
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
		return fmt.Errorf("can't update remote network : %w", err)
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.ok").Data().(bool)
	if !status {
		errorMessage := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.error").Data().(string)

		return APIError("can't update network: %s", errorMessage)
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
		return fmt.Errorf("can't delete remote network : %w", err)
	}

	status := deleteRemoteNetwork.Path("data.remoteNetworkDelete.ok").Data().(bool)
	if !status {
		errorMessage := deleteRemoteNetwork.Path("data.remoteNetworkDelete.error").Data().(string)

		return APIError("unable to delete network with Id %s, error: %s", remoteNetworkId, errorMessage)
	}

	return nil
}
