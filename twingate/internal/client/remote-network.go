package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type RemoteNetworkLocation string

func (client *Client) CreateRemoteNetwork(ctx context.Context, req *model.RemoteNetwork) (*model.RemoteNetwork, error) {
	opr := resourceRemoteNetwork.create()

	if req.Name == "" {
		return nil, opr.apiError(ErrGraphqlNetworkNameIsEmpty)
	}

	variables := newVars(
		gqlVar(req.Name, "name"),
		gqlVar(true, "isActive"),
		gqlVar(RemoteNetworkLocation(req.Location), "location"),
	)

	response := query.CreateRemoteNetwork{}
	if err := client.mutate(ctx, &response, variables, opr, attr{name: req.Name}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadRemoteNetworks(ctx context.Context) ([]*model.RemoteNetwork, error) {
	opr := resourceRemoteNetwork.read()

	variables := newVars(
		cursor(query.CursorRemoteNetworks),
		pageLimit(client.pageLimit),
	)

	response := query.ReadRemoteNetworks{}
	if err := client.query(ctx, &response, variables, opr.withCustomName("readRemoteNetworks"), attr{id: "All"}); err != nil {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readRemoteNetworksAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readRemoteNetworksAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.RemoteNetworkEdge], error) {
	opr := resourceRemoteNetwork.read()

	variables[query.CursorRemoteNetworks] = cursor

	response := query.ReadRemoteNetworks{}
	if err := client.query(ctx, &response, variables, opr.withCustomName("readRemoteNetworks")); err != nil {
		return nil, err
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
	opr := resourceRemoteNetwork.read()

	if remoteNetworkID == "" {
		return nil, opr.apiError(ErrGraphqlNetworkIDIsEmpty)
	}

	response := query.ReadRemoteNetworkByID{}
	if err := client.query(ctx, &response, newVars(gqlID(remoteNetworkID)),
		opr.withCustomName("readRemoteNetworkByID"), attr{id: remoteNetworkID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadRemoteNetworkByName(ctx context.Context, remoteNetworkName string) (*model.RemoteNetwork, error) {
	opr := resourceRemoteNetwork.read()

	if remoteNetworkName == "" {
		return nil, opr.apiError(ErrGraphqlNetworkNameIsEmpty)
	}

	response := query.ReadRemoteNetworkByName{}
	if err := client.query(ctx, &response, newVars(gqlVar(remoteNetworkName, "name")),
		opr.withCustomName("readRemoteNetworkByName"), attr{name: remoteNetworkName}); err != nil {
		return nil, err
	}

	return response.RemoteNetworks.Edges[0].Node.ToModel(), nil
}

func (client *Client) UpdateRemoteNetwork(ctx context.Context, req *model.RemoteNetwork) (*model.RemoteNetwork, error) {
	opr := resourceRemoteNetwork.update()

	variables := newVars(
		gqlID(req.ID),
		gqlNullable(req.Name, "name"),
		gqlVar(RemoteNetworkLocation(req.Location), "location"),
	)

	response := query.UpdateRemoteNetwork{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: req.ID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteRemoteNetwork(ctx context.Context, remoteNetworkID string) error {
	opr := resourceRemoteNetwork.delete()

	if remoteNetworkID == "" {
		return opr.apiError(ErrGraphqlNetworkIDIsEmpty)
	}

	response := query.DeleteRemoteNetwork{}

	return client.mutate(ctx, &response, newVars(gqlID(remoteNetworkID)), opr, attr{id: remoteNetworkID})
}
