package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const remoteNetworkResourceName = "remote network"

type gqlRemoteNetwork struct {
	ID   graphql.ID
	Name graphql.String
}

type gqlRemoteNetworkEdge struct {
	Node gqlRemoteNetwork
}

type gqlRemoteNetworks struct {
	Edges []*gqlRemoteNetworkEdge
}

type RemoteNetworks struct {
	PaginatedResource[*gqlRemoteNetworkEdge]
}

type createRemoteNetworkQuery struct {
	RemoteNetworkCreate struct {
		OkError
		Entity *gqlRemoteNetwork
	} `graphql:"remoteNetworkCreate(name: $name, isActive: $isActive)"`
}

func (client *Client) CreateRemoteNetwork(ctx context.Context, remoteNetworkName string) (*model.RemoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "create", remoteNetworkResourceName)
	}

	response := createRemoteNetworkQuery{}
	variables := newVars(
		gqlField(remoteNetworkName, "name"),
		gqlField(true, "isActive"),
	)

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

	return response.ToModel(), nil
}

type readRemoteNetworksQuery struct {
	RemoteNetworks RemoteNetworks
}

func (client *Client) ReadRemoteNetworks(ctx context.Context) ([]*model.RemoteNetwork, error) {
	response := readRemoteNetworksQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworks", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	err = response.RemoteNetworks.fetchPages(ctx, client.readRemoteNetworksAfter, nil)
	if err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

type readRemoteNetworksAfterQuery struct {
	RemoteNetworks RemoteNetworks `graphql:"remoteNetworks(after: $remoteNetworksEndCursor)"`
}

func (client *Client) readRemoteNetworksAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*gqlRemoteNetworkEdge], error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	variables["remoteNetworksEndCursor"] = cursor
	response := readRemoteNetworksAfterQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworks", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "read", remoteNetworkResourceName)
	}

	if len(response.RemoteNetworks.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName)
	}

	return &response.RemoteNetworks.PaginatedResource, nil
}

func (client *Client) ReadRemoteNetwork(ctx context.Context, remoteNetworkID, remoteNetworkName string) (*model.RemoteNetwork, error) {
	switch {
	case remoteNetworkID != "":
		return client.ReadRemoteNetworkByID(ctx, remoteNetworkID)
	default:
		return client.ReadRemoteNetworkByName(ctx, remoteNetworkName)
	}
}

type readRemoteNetworkByIDQuery struct {
	RemoteNetwork *gqlRemoteNetwork `graphql:"remoteNetwork(id: $id)"`
}

func (client *Client) ReadRemoteNetworkByID(ctx context.Context, remoteNetworkID string) (*model.RemoteNetwork, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := newVars(gqlID(remoteNetworkID))
	response := readRemoteNetworkByIDQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworkByID", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if response.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	return response.RemoteNetwork.ToModel(), nil
}

type readRemoteNetworkByNameQuery struct {
	RemoteNetworks gqlRemoteNetworks `graphql:"remoteNetworks(filter: {name: {eq: $name}})"`
}

func (client *Client) ReadRemoteNetworkByName(ctx context.Context, remoteNetworkName string) (*model.RemoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := newVars(gqlField(remoteNetworkName, "name"))
	response := readRemoteNetworkByNameQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readRemoteNetworkByName", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	if len(response.RemoteNetworks.Edges) == 0 || response.RemoteNetworks.Edges[0] == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	return response.RemoteNetworks.Edges[0].Node.ToModel(), nil
}

type updateRemoteNetworkQuery struct {
	RemoteNetworkUpdate struct {
		OkError
		Entity *gqlRemoteNetwork
	} `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (client *Client) UpdateRemoteNetwork(ctx context.Context, remoteNetworkID, remoteNetworkName string) (*model.RemoteNetwork, error) {
	variables := newVars(
		gqlID(remoteNetworkID),
		gqlField(remoteNetworkName, "name"),
	)

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

	return response.RemoteNetworkUpdate.Entity.ToModel(), nil
}

type deleteRemoteNetworkQuery struct {
	RemoteNetworkDelete *OkError `graphql:"remoteNetworkDelete(id: $id)"`
}

func (client *Client) DeleteRemoteNetwork(ctx context.Context, remoteNetworkID string) error {
	if remoteNetworkID == "" {
		return NewAPIError(ErrGraphqlNetworkIDIsEmpty, "delete", remoteNetworkResourceName)
	}

	variables := newVars(
		gqlID(remoteNetworkID),
	)
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
