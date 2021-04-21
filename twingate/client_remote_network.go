package twingate

import (
	"fmt"
)

type RemoteNetwork struct {
	Id   string
	Name string
}

func (client *Client) createRemoteNetwork(remoteNetworkName *string) (*RemoteNetwork, error) {
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
        `, *remoteNetworkName),
	}
	mutationRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, err
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkCreate.ok").Data().(bool)
	if !status {
		errorMessage := mutationRemoteNetwork.Path("data.remoteNetworkCreate.error").Data().(string)

		return nil, fmt.Errorf("cant create network with name %s, error: %w ", *remoteNetworkName, APIError(fmt.Sprintf("CreateNetwork failed with error : %s", errorMessage)))
	}

	remoteNetwork := RemoteNetwork{
		Id: mutationRemoteNetwork.Path("data.remoteNetworkCreate.entity.id").Data().(string),
	}

	return &remoteNetwork, nil
}

func (client *Client) readRemoteNetwork(remoteNetworkId *string) (*RemoteNetwork, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  remoteNetwork(id: "%s") {
			name
		  }
		}

        `, *remoteNetworkId),
	}
	queryRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, err
	}

	remoteNetworkQuery := queryRemoteNetwork.Path("data.remoteNetwork")

	if remoteNetworkQuery.Data() == nil {
		return nil, APIError(fmt.Sprintf("unable to read remote network :  %s", *remoteNetworkId))
	}

	remoteNetwork := RemoteNetwork{
		Id:   *remoteNetworkId,
		Name: remoteNetworkQuery.Path("name").Data().(string),
	}

	return &remoteNetwork, nil
}

func (client *Client) updateRemoteNetwork(remoteNetworkId, remoteNetworkName *string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
				mutation {
					remoteNetworkUpdate(id: "%s", name: "%s"){
						ok
						error
					}
				}
        `, *remoteNetworkId, *remoteNetworkName),
	}
	mutationRemoteNetwork, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return err
	}

	status := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.ok").Data().(bool)
	if !status {
		errorMessage := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.error").Data().(string)

		return fmt.Errorf("unable to update network:  %w", APIError(errorMessage))
	}

	return nil
}

func (client *Client) deleteRemoteNetwork(remoteNetworkId *string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  remoteNetworkDelete(id: "%s"){
			ok
			error
		  }
		}
		`, *remoteNetworkId),
	}
	deleteRemoteNetwork, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return err
	}

	status := deleteRemoteNetwork.Path("data.remoteNetworkDelete.ok").Data().(bool)
	if !status {
		errorMessage := deleteRemoteNetwork.Path("data.remoteNetworkDelete.error").Data().(string)

		return fmt.Errorf("unable to delete network with Id %s, error:  %w", *remoteNetworkId, APIError(errorMessage))
	}

	return nil
}
