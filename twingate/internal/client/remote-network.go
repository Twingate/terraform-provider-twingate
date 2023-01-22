package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

const remoteNetworkResourceName = "remote network"

type RemoteNetworkLocation graphql.String

func (client *Client) CreateRemoteNetwork(ctx context.Context, req *model.RemoteNetwork) (*model.RemoteNetwork, error) {
	if req.Name == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "create", remoteNetworkResourceName)
	}

	response := query.CreateRemoteNetwork{}
	variables := newVars(
		gqlVar(req.Name, "name"),
		gqlVar(true, "isActive"),
		gqlVar(RemoteNetworkLocation(req.Location), "location"),
	)

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createRemoteNetwork"))
	if err != nil {
		return nil, NewAPIError(err, "create", remoteNetworkResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", remoteNetworkResourceName)
	}

	if response.Entity == nil {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "create", remoteNetworkResourceName)
	}

	return response.ToModel(), nil
}

func (client *Client) ReadRemoteNetworks(ctx context.Context) ([]*model.RemoteNetwork, error) {
	response := query.ReadRemoteNetworks{}

	variables := newVars(gqlNullable("", query.CursorRemoteNetworks))

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readRemoteNetworks"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readRemoteNetworksAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readRemoteNetworksAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.RemoteNetworkEdge], error) {
	variables[query.CursorRemoteNetworks] = cursor
	response := query.ReadRemoteNetworks{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readRemoteNetworks"))
	if err != nil {
		return nil, NewAPIError(err, "read", remoteNetworkResourceName)
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName)
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadRemoteNetwork(ctx context.Context, remoteNetworkID, remoteNetworkName string) (*model.RemoteNetwork, error) {
	switch {
	case remoteNetworkID != "":
		return client.ReadRemoteNetworkByID(ctx, remoteNetworkID)
	default:
		return client.ReadRemoteNetworkByName(ctx, remoteNetworkName)
	}
}

func (client *Client) ReadRemoteNetworkByID(ctx context.Context, remoteNetworkID string) (*model.RemoteNetwork, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := newVars(gqlID(remoteNetworkID))
	response := query.ReadRemoteNetworkByID{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readRemoteNetworkByID"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	if response.RemoteNetwork == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkID)
	}

	return response.ToModel(), nil
}

func (client *Client) ReadRemoteNetworkByName(ctx context.Context, remoteNetworkName string) (*model.RemoteNetwork, error) {
	if remoteNetworkName == "" {
		return nil, NewAPIError(ErrGraphqlNetworkNameIsEmpty, "read", remoteNetworkResourceName)
	}

	variables := newVars(gqlVar(remoteNetworkName, "name"))
	response := query.ReadRemoteNetworkByName{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readRemoteNetworkByName"))
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	if len(response.RemoteNetworks.Edges) == 0 || response.RemoteNetworks.Edges[0] == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", remoteNetworkResourceName, remoteNetworkName)
	}

	return response.RemoteNetworks.Edges[0].Node.ToModel(), nil
}

func (client *Client) UpdateRemoteNetwork(ctx context.Context, req *model.RemoteNetwork) (*model.RemoteNetwork, error) {
	variables := newVars(
		gqlID(req.ID),
		gqlNullable(req.Name, "name"),
		gqlVar(RemoteNetworkLocation(req.Location), "location"),
	)

	response := query.UpdateRemoteNetwork{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateRemoteNetwork"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", remoteNetworkResourceName, req.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", remoteNetworkResourceName, req.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", remoteNetworkResourceName, req.ID)
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteRemoteNetwork(ctx context.Context, remoteNetworkID string) error {
	if remoteNetworkID == "" {
		return NewAPIError(ErrGraphqlNetworkIDIsEmpty, "delete", remoteNetworkResourceName)
	}

	variables := newVars(gqlID(remoteNetworkID))
	response := query.DeleteRemoteNetwork{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("deleteRemoteNetwork"))
	if err != nil {
		return NewAPIErrorWithID(err, "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "delete", remoteNetworkResourceName, remoteNetworkID)
	}

	return nil
}
