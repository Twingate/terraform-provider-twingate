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

func (client *Client) createRemoteNetwork(ctx context.Context, remoteNetworkName graphql.String) (*remoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "create", remoteNetworkResourceName)
	}

	response := createRemoteNetworkQuery{}

	variables := map[string]interface{}{
		"name":     remoteNetworkName,
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

	var remoteNetworks = make(map[int]*remoteNetwork)

	for i, elem := range response.RemoteNetworks.Edges {
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

func (client *Client) readRemoteNetwork(ctx context.Context, remoteNetworkID graphql.ID) (*remoteNetwork, error) {
	if remoteNetworkID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"id": remoteNetworkID,
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

func (client *Client) readRemoteNetworkByName(ctx context.Context, remoteNetworkName string) (*remoteNetwork, error) {
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

	if len(response.RemoteNetworks.Edges) > 1 {
		return nil, NewAPIErrorWithName(ErrGraphqlFoundMoreThanOneResource, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	network := response.RemoteNetworks.Edges[0]
	if network == nil || network.Node == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	return &remoteNetwork{
		ID:   network.Node.ID,
		Name: network.Node.Name,
	}, nil
}

type updateRemoteNetworkQuery struct {
	RemoteNetworkUpdate *OkError `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (client *Client) updateRemoteNetwork(ctx context.Context, remoteNetworkID graphql.ID, remoteNetworkName graphql.String) error {
	variables := map[string]interface{}{
		"id":   remoteNetworkID,
		"name": remoteNetworkName,
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

func (client *Client) deleteRemoteNetwork(ctx context.Context, remoteNetworkID graphql.ID) error {
	if remoteNetworkID.(string) == "" {
		return NewAPIError(ErrGraphqlNetworkIDIsEmpty, "delete", remoteNetworkResourceName)
	}

	variables := map[string]interface{}{
		"id": remoteNetworkID,
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
