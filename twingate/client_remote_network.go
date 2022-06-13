package twingate

import (
	"context"

	"github.com/twingate/go-graphql-client"
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

func (client *Client) createRemoteNetwork(ctx context.Context, remoteNetworkName string) (*remoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "create", remoteNetworkResourceName)
	}

	response := createRemoteNetworkQuery{}

	variables := map[string]interface{}{
		"name":     graphql.String(remoteNetworkName),
		"isActive": graphql.Boolean(true),
	}
	err := client.GraphqlClient.NamedMutate(ctx, "createRemoteNetwork", &response, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", remoteNetworkResourceName)
	}

	if !response.RemoteNetworkCreate.Ok {
		message := response.RemoteNetworkCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", remoteNetworkResourceName)
	}

	return &remoteNetwork{
		ID: response.RemoteNetworkCreate.Entity.ID,
	}, nil
}

type readRemoteNetworksQuery struct { //nolint
	RemoteNetworks struct {
		Edges []*Edges
	}
}

func (client *Client) readRemoteNetworks(ctx context.Context) (map[int]*remoteNetwork, error) { //nolint
	response := readRemoteNetworksQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworks", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	return convertRemoteNetworkResults(response.RemoteNetworks.Edges), nil
}

func convertRemoteNetworkResults(edges []*Edges) map[int]*remoteNetwork {
	var remoteNetworks = make(map[int]*remoteNetwork)

	for id, elem := range edges {
		remoteNetworks[id] = &remoteNetwork{
			ID:   elem.Node.StringID(),
			Name: elem.Node.Name,
		}
	}

	return remoteNetworks
}

type readRemoteNetworkQuery struct {
	RemoteNetwork *struct {
		Name graphql.String `json:"name"`
	} `graphql:"remoteNetwork(id: $id)"`
}

func (client *Client) readRemoteNetwork(ctx context.Context, remoteNetworkID string) (*remoteNetwork, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(remoteNetworkID),
	}

	response := readRemoteNetworkQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetwork", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if response.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	return &remoteNetwork{
		ID:   remoteNetworkID,
		Name: response.RemoteNetwork.Name,
	}, nil
}

type readRemoteNetworksByNameQuery struct {
	RemoteNetworks struct {
		Edges []*Edges
	} `graphql:"remoteNetworks(filter: {name: {eq: $name}})"`
}

func (client *Client) readRemoteNetworksByName(ctx context.Context, remoteNetworkName string) (map[int]*remoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"name": graphql.String(remoteNetworkName),
	}

	response := readRemoteNetworksByNameQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworkByName", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	if len(response.RemoteNetworks.Edges) == 0 {
		return nil, NewAPIErrorWithName(ErrGraphqlResourceNotFound, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	return convertRemoteNetworkResults(response.RemoteNetworks.Edges), nil
}

type updateRemoteNetworkQuery struct {
	RemoteNetworkUpdate *OkError `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (client *Client) updateRemoteNetwork(ctx context.Context, remoteNetworkID, remoteNetworkName string) error {
	variables := map[string]interface{}{
		"id":   graphql.ID(remoteNetworkID),
		"name": graphql.String(remoteNetworkName),
	}

	response := updateRemoteNetworkQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateRemoteNetwork", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "update", remoteNetworkResourceName, remoteNetworkID)
	}

	if !response.RemoteNetworkUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.RemoteNetworkUpdate.Error), "update", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}

type deleteRemoteNetworkQuery struct {
	RemoteNetworkDelete *OkError `graphql:"remoteNetworkDelete(id: $id)"`
}

func (client *Client) deleteRemoteNetwork(ctx context.Context, remoteNetworkID string) error {
	if remoteNetworkID == "" {
		return NewAPIError(ErrGraphqlNetworkIDIsEmpty, "delete", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(remoteNetworkID),
	}

	response := deleteRemoteNetworkQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteRemoteNetwork", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	if !response.RemoteNetworkDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.RemoteNetworkDelete.Error), "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}
