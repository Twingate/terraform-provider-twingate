package transport

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const connectorResourceName = "connector"

type remoteNetwork struct {
	ID   graphql.ID
	Name graphql.String
}

type gqlConnector struct {
	IDName
	RemoteNetwork struct {
		ID graphql.ID
	}
}

type createConnectorQuery struct {
	ConnectorCreate struct {
		Entity *gqlConnector
		OkError
	} `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId, name: $connectorName)"`
}

func (client *Client) CreateConnector(ctx context.Context, remoteNetworkID, connectorName string) (*model.Connector, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	variables := newVars(
		gqlID(remoteNetworkID, "remoteNetworkId"),
		gqlNullableField(connectorName, "connectorName"),
	)

	response := createConnectorQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createConnector", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	if !response.ConnectorCreate.Ok {
		return nil, NewAPIError(NewMutationError(response.ConnectorCreate.Error), "create", connectorResourceName)
	}

	if response.ConnectorCreate.Entity == nil {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "create", connectorResourceName)
	}

	return response.ConnectorCreate.Entity.ToModel(), nil
}

type updateConnectorQuery struct {
	ConnectorUpdate struct {
		Entity *gqlConnector
		OkError
	} `graphql:"connectorUpdate(id: $connectorId, name: $connectorName)"`
}

func (client *Client) UpdateConnector(ctx context.Context, connectorID string, connectorName string) (*model.Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
	}

	variables := newVars(
		gqlID(connectorID, "connectorId"),
		gqlField(connectorName, "connectorName"),
	)

	response := updateConnectorQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateConnector", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", connectorResourceName, connectorID)
	}

	if !response.ConnectorUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.ConnectorUpdate.Error), "update", connectorResourceName, connectorID)
	}

	if response.ConnectorUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", connectorResourceName, connectorID)
	}

	return response.ConnectorUpdate.Entity.ToModel(), nil
}

type deleteConnectorQuery struct {
	ConnectorDelete *OkError `graphql:"connectorDelete(id: $id)"`
}

func (client *Client) DeleteConnector(ctx context.Context, connectorID string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", connectorResourceName)
	}

	variables := newVars(gqlID(connectorID))
	response := deleteConnectorQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteConnector", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	if !response.ConnectorDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ConnectorDelete.Error), "delete", connectorResourceName, connectorID)
	}

	return nil
}

type readConnectorQuery struct {
	Connector *gqlConnector `graphql:"connector(id: $id)"`
}

func (client *Client) ReadConnector(ctx context.Context, connectorID string) (*model.Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", connectorResourceName)
	}

	variables := newVars(gqlID(connectorID))
	response := readConnectorQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnector", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if response.Connector == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, connectorID)
	}

	return response.ToModel(), nil
}

type ConnectorEdge struct {
	Node *gqlConnector
}

type Connectors struct {
	PageInfo PageInfo
	Edges    []*ConnectorEdge
}

type readConnectorsQuery struct {
	Connectors Connectors
}

func (client *Client) ReadConnectors(ctx context.Context) ([]*model.Connector, error) {
	response := readConnectorsQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnectors", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if len(response.Connectors.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	connectors, err := client.readAllConnectors(ctx, &response.Connectors)
	if err != nil {
		return nil, err
	}

	return connectors.ToModel(), nil
}

func (client *Client) readAllConnectors(ctx context.Context, connectors *Connectors) (*Connectors, error) {
	page := connectors.PageInfo
	for page.HasNextPage {
		resp, err := client.readConnectorsAfter(ctx, page.EndCursor)
		if err != nil {
			return nil, err
		}

		connectors.Edges = append(connectors.Edges, resp.Edges...)
		page = resp.PageInfo
	}

	return connectors, nil
}

type readConnectorsAfter struct {
	Connectors Connectors `graphql:"connectors(after: $connectorsEndCursor)"`
}

func (client *Client) readConnectorsAfter(ctx context.Context, cursor graphql.String) (*Connectors, error) {
	response := readConnectorsAfter{}
	variables := map[string]interface{}{
		"connectorsEndCursor": cursor,
	}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnectors", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if response.Connectors.Edges == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	return &response.Connectors, nil
}