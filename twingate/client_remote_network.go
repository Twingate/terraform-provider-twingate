package twingate

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type remoteNetwork struct {
	ID   graphql.ID
	Name graphql.String
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

func (client *Client) createRemoteNetwork(remoteNetworkName graphql.String) (*remoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "create", remoteNetworkResourceName, "remoteNetworkName")
	}

	r := createRemoteNetworkQuery{}

	variables := map[string]interface{}{
		"name":     remoteNetworkName,
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
		ID: r.RemoteNetworkCreate.Entity.ID,
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
		c := &remoteNetwork{ID: elem.Node.StringID(), Name: elem.Node.Name}
		remoteNetworks[i] = c
	}

	return remoteNetworks, nil
}

type readRemoteNetworkQuery struct {
	RemoteNetwork *struct {
		Name graphql.String `json:"name"`
	} `graphql:"remoteNetwork(id: $id)"`
}

func (client *Client) readRemoteNetwork(remoteNetworkID graphql.ID) (*remoteNetwork, error) {
	if remoteNetworkID == nil || remoteNetworkID == "" {
		return nil, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "read", remoteNetworkResourceName, "remoteNetworkID")
	}

	variables := map[string]interface{}{
		"id": remoteNetworkID,
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
		Name: r.RemoteNetwork.Name,
	}, nil
}

type updateRemoteNetworkQuery struct {
	RemoteNetworkUpdate *OkError `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (client *Client) updateRemoteNetwork(remoteNetworkID graphql.ID, remoteNetworkName graphql.String) error {
	variables := map[string]interface{}{
		"id":   remoteNetworkID,
		"name": remoteNetworkName,
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

func (client *Client) deleteRemoteNetwork(remoteNetworkID graphql.ID) error {
	if remoteNetworkID == nil || remoteNetworkID == "" {
		return NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "delete", remoteNetworkResourceName, "remoteNetworkID")
	}

	variables := map[string]interface{}{
		"id": remoteNetworkID,
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
