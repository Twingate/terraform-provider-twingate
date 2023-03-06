package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

const connectorResourceName = "connector"

func (client *Client) CreateConnector(ctx context.Context, input *model.Connector) (*model.Connector, error) {
	if input == nil || input.NetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
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

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createConnector"))
	if err != nil {
		return nil, NewAPIErrorWithName(err, "create", connectorResourceName, input.Name)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithName(NewMutationError(response.Error), "create", connectorResourceName, input.Name)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "create", connectorResourceName, input.Name)
	}

	return response.Entity.ToModel(), nil
}

func (client *Client) UpdateConnector(ctx context.Context, input *model.Connector) (*model.Connector, error) {
	if input == nil || input.ID == "" {
		return nil, NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
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

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateConnector"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", connectorResourceName, input.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", connectorResourceName, input.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", connectorResourceName, input.ID)
	}

	return response.Entity.ToModel(), nil
}

func (client *Client) DeleteConnector(ctx context.Context, connectorID string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", connectorResourceName)
	}

	variables := newVars(gqlID(connectorID))
	response := query.DeleteConnector{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("deleteConnector"))
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "delete", connectorResourceName, connectorID)
	}

	return nil
}

func (client *Client) ReadConnector(ctx context.Context, connectorID string) (*model.Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", connectorResourceName)
	}

	variables := newVars(gqlID(connectorID))
	response := query.ReadConnector{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readConnector"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if response.Connector == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, connectorID)
	}

	return response.ToModel(), nil
}

func (client *Client) ReadConnectors(ctx context.Context) ([]*model.Connector, error) {
	response := query.ReadConnectors{}
	variables := newVars(gqlNullable("", query.CursorConnectors))

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readConnectors"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readConnectorsAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readConnectorsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ConnectorEdge], error) {
	variables[query.CursorConnectors] = cursor
	response := query.ReadConnectors{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readConnectors"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	return &response.PaginatedResource, nil
}
