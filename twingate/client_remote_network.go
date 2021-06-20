package twingate

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type remoteNetwork struct {
	ID   string
	Name string
}

const remoteNetworkResourceName = "remote network"

type createRemoteNetworkQuery struct {
	RemoteNetworkCreate *struct {
		OkError
		Entity struct {
			ID graphql.ID
		}
	} `graphql:"remoteNetworkCreate(name: $name, isActive: $isActive)"`
}

func (client *Client) createRemoteNetwork(remoteNetworkName string) (*remoteNetwork, error) {
	r := createRemoteNetworkQuery{}

	variables := map[string]interface{}{
		"name":     graphql.String(remoteNetworkName),
		"isActive": graphql.Boolean(true),
	}
	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", remoteNetworkResourceName)
	}

	if !r.RemoteNetworkCreate.Ok {
		message := r.RemoteNetworkCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", remoteNetworkResourceName)
	}

	return &remoteNetwork{
		ID: r.RemoteNetworkCreate.Entity.ID.(string),
	}, nil
}

type readRemoteNetworksQuery struct { //nolint
	RemoteNetworks struct {
		Edges []*Edges
	}
}

func (client *Client) readRemoteNetworks() (map[int]*remoteNetwork, error) { //nolint
	r := readRemoteNetworksQuery{}

	err := client.GraphqlClient.Query(context.Background(), &r, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	var remoteNetworks = make(map[int]*remoteNetwork)

	for i, elem := range r.RemoteNetworks.Edges {
		c := &remoteNetwork{ID: elem.Node.StringID(), Name: elem.Node.StringName()}
		remoteNetworks[i] = c
	}

	return remoteNetworks, nil
}

type readRemoteNetworkQuery struct {
	RemoteNetwork *struct {
		Name graphql.String `json:"name"`
	} `graphql:"remoteNetwork(id: $id)"`
}

func (client *Client) readRemoteNetwork(remoteNetworkID string) (*remoteNetwork, error) {
	variables := map[string]interface{}{
		"id": graphql.ID(remoteNetworkID),
	}

	r := readRemoteNetworkQuery{}

	err := client.GraphqlClient.Query(context.Background(), &r, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if r.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	return &remoteNetwork{
		ID:   remoteNetworkID,
		Name: string(r.RemoteNetwork.Name),
	}, nil
}

type updateRemoteNetworkQuery struct {
	RemoteNetworkUpdate *OkError `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (client *Client) updateRemoteNetwork(remoteNetworkID, remoteNetworkName string) error {
	variables := map[string]interface{}{
		"id":   graphql.ID(remoteNetworkID),
		"name": graphql.String(remoteNetworkName),
	}

	r := updateRemoteNetworkQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "update", remoteNetworkResourceName, remoteNetworkID)
	}

	if !r.RemoteNetworkUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(r.RemoteNetworkUpdate.Error), "update", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}

type deleteRemoteNetworkQuery struct {
	RemoteNetworkDelete *OkError `graphql:"remoteNetworkDelete(id: $id)"`
}

func (client *Client) deleteRemoteNetwork(remoteNetworkID string) error {
	variables := map[string]interface{}{
		"id": graphql.ID(remoteNetworkID),
	}

	r := deleteRemoteNetworkQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	if !r.RemoteNetworkDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(r.RemoteNetworkDelete.Error), "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}
