package twingate

import (
	"fmt"
)

type RemoteNetwork struct {
	ID   string
	Name string
}

const remoteNetworkResourceName = "remote network"

type createRemoteNetworkResponse struct {
	Data *createRemoteNetworkResponseData `json:"data"`
}

type createRemoteNetworkResponseData struct {
	RemoteNetworkCreate *createRemoteNetworkResponseDataRemoteNetworkCreate `json:"remoteNetworkCreate"`
}

type createRemoteNetworkResponseDataRemoteNetworkCreate struct {
	Ok     bool                                                      `json:"ok"`
	Error  string                                                    `json:"error"`
	Entity *createRemoteNetworkResponseDataRemoteNetworkCreateEntity `json:"entity"`
}

type createRemoteNetworkResponseDataRemoteNetworkCreateEntity struct {
	Id string `json:"id"`
}

func newCreateRemoteNetworkResponse() *createRemoteNetworkResponse {
	return &createRemoteNetworkResponse{
		Data: &createRemoteNetworkResponseData{
			RemoteNetworkCreate: &createRemoteNetworkResponseDataRemoteNetworkCreate{
				Entity: &createRemoteNetworkResponseDataRemoteNetworkCreateEntity{},
			},
		},
	}
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

	r := newCreateRemoteNetworkResponse()

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIError(err, "create", remoteNetworkResourceName)
	}

	if !r.Data.RemoteNetworkCreate.Ok {
		message := r.Data.RemoteNetworkCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", remoteNetworkResourceName)
	}

	remoteNetwork := RemoteNetwork{
		ID: r.Data.RemoteNetworkCreate.Entity.Id,
	}

	return &remoteNetwork, nil
}

type readRemoteNetworkResponse struct {
	Data *readRemoteNetworkResponseData `json:"data"`
}

type readRemoteNetworkResponseData struct {
	RemoteNetwork *readRemoteNetworkResponseDataRemoteNetwork `json:"remoteNetwork"`
}

type readRemoteNetworkResponseDataRemoteNetwork struct {
	Name string `json:"name"`
}

func newReadRemoteNetworkResponse() *readRemoteNetworkResponse {
	return &readRemoteNetworkResponse{
		Data: &readRemoteNetworkResponseData{
			RemoteNetwork: &readRemoteNetworkResponseDataRemoteNetwork{},
		},
	}
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

	r := newReadRemoteNetworkResponse()

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if r.Data == nil || r.Data.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	remoteNetwork := RemoteNetwork{
		ID:   remoteNetworkID,
		Name: r.Data.RemoteNetwork.Name,
	}

	return &remoteNetwork, nil
}

type updateRemoteNetworkResponse struct {
	Data struct {
		RemoteNetworkUpdate struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		} `json:"remoteNetworkUpdate"`
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
	Data *deleteRemoteNetworkResponseData `json:"data"`
}

type deleteRemoteNetworkResponseData struct {
	RemoteNetworkDelete *deleteRemoteNetworkResponseDataRemoteNetworkDelete `json:"remoteNetworkDelete"`
}

type deleteRemoteNetworkResponseDataRemoteNetworkDelete struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func newDeleteRemoteNetworkResponse() *deleteRemoteNetworkResponse {
	return &deleteRemoteNetworkResponse{
		Data: &deleteRemoteNetworkResponseData{
			RemoteNetworkDelete: &deleteRemoteNetworkResponseDataRemoteNetworkDelete{},
		},
	}
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

	r := newDeleteRemoteNetworkResponse()

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
