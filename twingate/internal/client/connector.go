package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

const connectorResourceName = "connector"

func (client *Client) CreateConnector(ctx context.Context, remoteNetworkID, connectorName string) (*model.Connector, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	variables := newVars(
		gqlID(remoteNetworkID, "remoteNetworkId"),
		gqlNullable(connectorName, "connectorName"),
	)

	response := query.CreateConnector{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createConnector"))
	if err != nil {
		return nil, NewAPIErrorWithName(err, "create", connectorResourceName, connectorName)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithName(NewMutationError(response.Error), "create", connectorResourceName, connectorName)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "create", connectorResourceName, connectorName)
	}

	return response.Entity.ToModel(), nil
}

func (client *Client) UpdateConnector(ctx context.Context, connectorID string, connectorName string) (*model.Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
	}

	variables := newVars(
		gqlID(connectorID, "connectorId"),
		gqlVar(connectorName, "connectorName"),
	)

	response := query.UpdateConnector{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateConnector"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", connectorResourceName, connectorID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", connectorResourceName, connectorID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", connectorResourceName, connectorID)
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

func (client *Client) readConnectorsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.ConnectorEdge], error) {
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
