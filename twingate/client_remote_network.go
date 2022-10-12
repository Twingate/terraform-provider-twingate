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
		Entity *remoteNetwork
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

	if response.RemoteNetworkCreate.Entity == nil {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "create", remoteNetworkResourceName)
	}

	return response.RemoteNetworkCreate.Entity, nil
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

	var remoteNetworks = make(map[int]*remoteNetwork)

	for i, elem := range response.RemoteNetworks.Edges {
		c := &remoteNetwork{ID: elem.Node.StringID(), Name: elem.Node.Name}
		remoteNetworks[i] = c
	}

	return remoteNetworks, nil
}

func (client *Client) readRemoteNetwork(ctx context.Context, remoteNetworkID, remoteNetworkName string) (*remoteNetwork, error) {
	switch {
	case remoteNetworkID != "":
		return client.readRemoteNetworkByID(ctx, remoteNetworkID)
	default:
		return client.readRemoteNetworkByName(ctx, remoteNetworkName)
	}
}

type readRemoteNetworkByIDQuery struct {
	RemoteNetwork *remoteNetwork `graphql:"remoteNetwork(id: $id)"`
}

func (client *Client) readRemoteNetworkByID(ctx context.Context, remoteNetworkID string) (*remoteNetwork, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(remoteNetworkID),
	}

	response := readRemoteNetworkByIDQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworkByID", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if response.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	return response.RemoteNetwork, nil
}

type readRemoteNetworkByNameQuery struct {
	RemoteNetworks struct {
		Edges []*Edges
	} `graphql:"remoteNetworks(filter: {name: {eq: $name}})"`
}

func (client *Client) readRemoteNetworkByName(ctx context.Context, remoteNetworkName string) (*remoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"name": graphql.String(remoteNetworkName),
	}

	response := readRemoteNetworkByNameQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworkByName", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	if len(response.RemoteNetworks.Edges) == 0 || response.RemoteNetworks.Edges[0].Node == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	node := response.RemoteNetworks.Edges[0].Node

	return &remoteNetwork{
		ID:   node.ID,
		Name: node.Name,
	}, nil
}

type updateRemoteNetworkQuery struct {
	RemoteNetworkUpdate struct {
		OkError
		Entity *remoteNetwork
	} `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (client *Client) updateRemoteNetwork(ctx context.Context, remoteNetworkID, remoteNetworkName string) (*remoteNetwork, error) {
	variables := map[string]interface{}{
		"id":   graphql.ID(remoteNetworkID),
		"name": graphql.String(remoteNetworkName),
	}

	response := updateRemoteNetworkQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateRemoteNetwork", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", remoteNetworkResourceName, remoteNetworkID)
	}

	if !response.RemoteNetworkUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.RemoteNetworkUpdate.Error), "update", remoteNetworkResourceName, remoteNetworkID)
	}

	if response.RemoteNetworkUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", remoteNetworkResourceName, remoteNetworkID)
	}

	return response.RemoteNetworkUpdate.Entity, nil
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
