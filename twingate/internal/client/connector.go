package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

func (client *Client) CreateConnector(ctx context.Context, input *model.Connector) (*model.Connector, error) {
	opr := resourceConnector.create()

	if input == nil || input.NetworkID == "" {
		return nil, opr.apiError(ErrGraphqlNetworkIDIsEmpty)
	}

	variables := newVars(
		gqlID(input.NetworkID, "remoteNetworkId"),
		gqlNullable(input.Name, "connectorName"),
	)

	if input.StatusUpdatesEnabled == nil {
		variables = gqlNullable(false, "hasStatusNotificationsEnabled")(variables)
	} else {
		variables = gqlVar(*input.StatusUpdatesEnabled, "hasStatusNotificationsEnabled")(variables)
	}

	var response query.CreateConnector
	if err := client.mutate(ctx, &response, variables, opr, attr{name: input.Name}); err != nil {
		return nil, err
	}

	return response.Entity.ToModel(), nil
}

func (client *Client) UpdateConnector(ctx context.Context, input *model.Connector) (*model.Connector, error) {
	opr := resourceConnector.update()

	if input == nil || input.ID == "" {
		return nil, opr.apiError(ErrGraphqlConnectorIDIsEmpty)
	}

	variables := newVars(
		gqlID(input.ID, "connectorId"),
		gqlNullable(input.Name, "connectorName"),
	)

	if input.StatusUpdatesEnabled == nil {
		variables = gqlNullable(false, "hasStatusNotificationsEnabled")(variables)
	} else {
		variables = gqlVar(*input.StatusUpdatesEnabled, "hasStatusNotificationsEnabled")(variables)
	}

	response := query.UpdateConnector{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	return response.Entity.ToModel(), nil
}

func (client *Client) DeleteConnector(ctx context.Context, connectorID string) error {
	opr := resourceConnector.delete()

	if connectorID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteConnector{}

	return client.mutate(ctx, &response, newVars(gqlID(connectorID)), opr, attr{id: connectorID})
}

func (client *Client) ReadConnector(ctx context.Context, connectorID string) (*model.Connector, error) {
	opr := resourceConnector.read()

	if connectorID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.ReadConnector{}
	if err := client.query(ctx, &response, newVars(gqlID(connectorID)), opr, attr{id: connectorID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadConnectors(ctx context.Context) ([]*model.Connector, error) {
	op := resourceConnector.read()

	variables := newVars(gqlNullable("", query.CursorConnectors))

	response := query.ReadConnectors{}
	if err := client.query(ctx, &response, variables, op.withCustomName("readConnectors"), attr{id: "All"}); err != nil {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readConnectorsAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readConnectorsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ConnectorEdge], error) {
	op := resourceConnector.read()

	variables[query.CursorConnectors] = cursor

	response := query.ReadConnectors{}
	if err := client.query(ctx, &response, variables, op.withCustomName("readConnectors"), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}
