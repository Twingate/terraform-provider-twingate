package twingate

import (
	"fmt"
)

type remoteNetwork struct {
	ID   string
	Name string
}

const remoteNetworkResourceName = "remote network"

type createRemoteNetworkResponse struct {
	Data *struct {
		RemoteNetworkCreate *struct {
			*OkErrorResponse
			Entity *struct {
				ID string `json:"id"`
			} `json:"entity"`
		} `json:"remoteNetworkCreate"`
	} `json:"data"`
}

func (client *Client) createRemoteNetwork(remoteNetworkName string) (*remoteNetwork, error) {
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

	r := createRemoteNetworkResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIError(err, "create", remoteNetworkResourceName)
	}

	if !r.Data.RemoteNetworkCreate.Ok {
		message := r.Data.RemoteNetworkCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", remoteNetworkResourceName)
	}

	remoteNetwork := remoteNetwork{
		ID: r.Data.RemoteNetworkCreate.Entity.ID,
	}

	return &remoteNetwork, nil
}

type readRemoteNetworksResponse struct {
	Data struct {
		RemoteNetworks struct {
			Edges []*EdgesResponse `json:"edges"`
		} `json:"remoteNetworks"`
	} `json:"data"`
}

func (client *Client) readRemoteNetworks() (map[int]*remoteNetwork, error) { //nolint
	query := map[string]string{
		"query": "{ remoteNetworks { edges { node { id name } } } }",
	}

	r := readRemoteNetworksResponse{}
	err := client.doGraphqlRequest(query, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	var remoteNetworks = make(map[int]*remoteNetwork)

	for i, elem := range r.Data.RemoteNetworks.Edges {
		c := &remoteNetwork{ID: elem.Node.ID, Name: elem.Node.Name}
		remoteNetworks[i] = c
	}

	return remoteNetworks, nil
}

type readRemoteNetworkResponse struct {
	Errors []*readRemoteNetworkResponseErrors `json:"errors"`
	Data   *struct {
		RemoteNetwork *struct {
			Name string `json:"name"`
		} `json:"remoteNetwork"`
	} `json:"data"`
}

type readRemoteNetworkResponseErrors struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Path []string `json:"path"`
}

func (r *readRemoteNetworkResponse) parseErrors() []string {
	var messages []string
	for _, e := range r.Errors {
		messages = append(messages, e.Message)
	}
	return messages
}

func (client *Client) readRemoteNetwork(remoteNetworkID string) (*remoteNetwork, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  remoteNetwork(id: "%s") {
			name
		  }
		}

        `, remoteNetworkID),
	}

	r := readRemoteNetworkResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if r.Errors != nil {
		return nil, NewAPIErrorWithID(NewGraphQLError(r.parseErrors()), "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if r.Data == nil || r.Data.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	remoteNetwork := remoteNetwork{
		ID:   remoteNetworkID,
		Name: r.Data.RemoteNetwork.Name,
	}

	return &remoteNetwork, nil
}

type updateRemoteNetworkResponse struct {
	Data struct {
		RemoteNetworkUpdate *OkErrorResponse `json:"remoteNetworkUpdate"`
	} `json:"data"`
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

	r := updateRemoteNetworkResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "update", remoteNetworkResourceName, remoteNetworkID)
	}

	if !r.Data.RemoteNetworkUpdate.Ok {
		message := r.Data.RemoteNetworkUpdate.Error

		return NewAPIErrorWithID(NewMutationError(message), "update", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}

type deleteRemoteNetworkResponse struct {
	Data *struct {
		RemoteNetworkDelete *OkErrorResponse `json:"remoteNetworkDelete"`
	} `json:"data"`
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

	r := deleteRemoteNetworkResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	if !r.Data.RemoteNetworkDelete.Ok {
		message := r.Data.RemoteNetworkDelete.Error

		return NewAPIErrorWithID(NewMutationError(message), "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}
