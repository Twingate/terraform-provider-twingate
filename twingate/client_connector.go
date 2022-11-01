package twingate

import (
	"context"

	"github.com/twingate/go-graphql-client"
)

type Connector struct {
	ID            graphql.ID
	RemoteNetwork *remoteNetwork
	Name          graphql.String
}

type Connectors struct {
	PageInfo struct {
		EndCursor   graphql.String
		HasNextPage graphql.Boolean
	}
	Edges []*struct {
		Node *Connector
	}
}

func (c *Connectors) toList() []*Connector {
	connectors := make([]*Connector, 0, len(c.Edges))
	for _, elem := range c.Edges {
		connectors = append(connectors, elem.Node)
	}

	return connectors
}

const connectorResourceName = "connector"

func (client *Client) createConnector(ctx context.Context, remoteNetworkID, connectorName string) (*Connector, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	var (
		err      error
		response *ConnectorCreateResponse
	)

	if connectorName == "" {
		response, err = client.createConnectorWithoutName(ctx, remoteNetworkID)
	} else {
		response, err = client.createConnectorWithName(ctx, remoteNetworkID, connectorName)
	}

	if err != nil {
		return nil, err
	}

	return response.Entity, nil
}

type ConnectorCreateResponse struct {
	Entity *Connector
	OkError
}

type createConnectorQuery struct {
	ConnectorCreate *ConnectorCreateResponse `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId)"`
}

func (client *Client) createConnectorWithoutName(ctx context.Context, remoteNetworkID string) (*ConnectorCreateResponse, error) {
	variables := map[string]interface{}{
		"remoteNetworkId": graphql.ID(remoteNetworkID),
	}

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

	return response.ConnectorCreate, nil
}

type createConnectorWithNameQuery struct {
	ConnectorCreate *ConnectorCreateResponse `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId, name: $connectorName)"`
}

func (client *Client) createConnectorWithName(ctx context.Context, remoteNetworkID, connectorName string) (*ConnectorCreateResponse, error) {
	variables := map[string]interface{}{
		"remoteNetworkId": graphql.ID(remoteNetworkID),
		"connectorName":   graphql.String(connectorName),
	}

	response := createConnectorWithNameQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createConnector", &response, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	if !response.ConnectorCreate.Ok {
		return nil, NewAPIError(NewMutationError(response.ConnectorCreate.Error), "create", connectorResourceName)
	}

	if response.ConnectorCreate.Entity == nil {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "create", connectorResourceName, connectorName)
	}

	return response.ConnectorCreate, nil
}

type updateConnectorQuery struct {
	ConnectorUpdate struct {
		Entity *Connector
		OkError
	} `graphql:"connectorUpdate(id: $connectorId, name: $connectorName )"`
}

func (client *Client) updateConnector(ctx context.Context, connectorID string, connectorName string) (*Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
	}

	variables := map[string]interface{}{
		"connectorId":   graphql.ID(connectorID),
		"connectorName": graphql.String(connectorName),
	}
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

	return response.ConnectorUpdate.Entity, nil
}

type readConnectorsQuery struct {
	Connectors Connectors
}

func (client *Client) readConnectors(ctx context.Context) ([]*Connector, error) {
	response := readConnectorsQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnectors", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if response.Connectors.Edges == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	connectors, err := client.readAllConnectors(ctx, &response.Connectors)
	if err != nil {
		return nil, err
	}

	return connectors.toList(), nil
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

type readConnectorQuery struct {
	Connector *Connector `graphql:"connector(id: $id)"`
}

func (client *Client) readConnector(ctx context.Context, connectorID string) (*Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", connectorResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(connectorID),
	}

	response := readConnectorQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnector", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if response.Connector == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, connectorID)
	}

	return response.Connector, nil
}

type deleteConnectorQuery struct {
	ConnectorDelete *OkError `graphql:"connectorDelete(id: $id)" json:"connectorDelete"`
}

func (client *Client) deleteConnector(ctx context.Context, connectorID string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", connectorResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(connectorID),
	}

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
