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
	Error *struct {
		Errors []*queryResponseErrors `json:"errors"`
	} `json:"error"`
	Data *struct {
		RemoteNetworkCreate *struct {
			*OkErrorResponse
			Entity *struct {
				ID string `json:"id"`
			} `json:"entity"`
		} `json:"remoteNetworkCreate"`
	} `json:"data"`
}

func (r *createRemoteNetworkResponse) checkErrors() []*queryResponseErrors {
	if r.Error != nil {
		return r.Error.Errors
	}

	return nil
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

	return &remoteNetwork{
		ID: r.Data.RemoteNetworkCreate.Entity.ID,
	}, nil
}

type readRemoteNetworksResponse struct { //nolint
	Error *struct {
		Errors []*queryResponseErrors `json:"errors"`
	} `json:"error"`
	Data struct {
		RemoteNetworks struct {
			Edges []*EdgesResponse `json:"edges"`
		} `json:"remoteNetworks"`
	} `json:"data"`
}

func (r *readRemoteNetworksResponse) checkErrors() []*queryResponseErrors { //nolint
	if r.Error != nil {
		return r.Error.Errors
	}

	return nil
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
	Errors []*queryResponseErrors `json:"errors"`
	Data   *struct {
		RemoteNetwork *struct {
			Name string `json:"name"`
		} `json:"remoteNetwork"`
	} `json:"data"`
}

func (r *readRemoteNetworkResponse) checkErrors() []*queryResponseErrors {
	return r.Errors
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

	if r.Data == nil || r.Data.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	return &remoteNetwork{
		ID:   remoteNetworkID,
		Name: r.Data.RemoteNetwork.Name,
	}, nil
}

type updateRemoteNetworkResponse struct {
	Data struct {
		RemoteNetworkUpdate *OkErrorResponse `json:"remoteNetworkUpdate"`
	} `json:"data"`
}

func (r *updateRemoteNetworkResponse) checkErrors() []*queryResponseErrors {
	return nil
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
		return NewAPIErrorWithID(NewMutationError(r.Data.RemoteNetworkUpdate.Error), "update", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}

type deleteRemoteNetworkResponse struct {
	Data *struct {
		RemoteNetworkDelete *OkErrorResponse `json:"remoteNetworkDelete"`
	} `json:"data"`
}

func (r *deleteRemoteNetworkResponse) checkErrors() []*queryResponseErrors {
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

	r := deleteRemoteNetworkResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	if !r.Data.RemoteNetworkDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(r.Data.RemoteNetworkDelete.Error), "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}
