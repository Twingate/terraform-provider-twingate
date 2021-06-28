package twingate

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Connector struct {
	ID              graphql.ID
	RemoteNetwork   *remoteNetwork
	Name            graphql.String
	ConnectorTokens *connectorTokens
}

type Connectors struct {
	ID   string
	Name string
}

const connectorResourceName = "connector"

type createConnectorQuery struct {
	ConnectorCreate struct {
		Entity IDName
		OkError
	} `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId)"`
}

func (client *Client) createConnector(remoteNetworkID graphql.ID) (*Connector, error) {
	if checkEmptyID(remoteNetworkID) {
		return nil, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "create", remoteNetworkResourceName, "remoteNetworkID")
	}

	variables := map[string]interface{}{
		"remoteNetworkId": remoteNetworkID,
	}
	r := createConnectorQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	if !r.ConnectorCreate.Ok {
		return nil, NewAPIError(NewMutationError(r.ConnectorCreate.Error), "create", connectorResourceName)
	}

	connector := Connector{
		ID:   r.ConnectorCreate.Entity.ID,
		Name: r.ConnectorCreate.Entity.Name,
	}

	return &connector, nil
}

type readConnectorsQuery struct { //nolint
	Connectors struct {
		Edges []*Edges
	}
}

func (client *Client) readConnectors() (map[int]*Connectors, error) { //nolint
	r := readConnectorsQuery{}

	err := client.GraphqlClient.Query(context.Background(), &r, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	var connectors = make(map[int]*Connectors)

	for i, elem := range r.Connectors.Edges {
		c := &Connectors{ID: elem.Node.StringID(), Name: elem.Node.StringName()}
		connectors[i] = c
	}

	return connectors, nil
}

type readConnectorQuery struct {
	Connector *struct {
		IDName
		RemoteNetwork IDName
	} `graphql:"connector(id: $id)"`
}

func (client *Client) readConnector(connectorID graphql.ID) (*Connector, error) {
	if checkEmptyID(connectorID) {
		return nil, NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "read", remoteNetworkResourceName, "connectorID")
	}

	variables := map[string]interface{}{
		"id": connectorID,
	}

	r := readConnectorQuery{}

	err := client.GraphqlClient.Query(context.Background(), &r, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if r.Connector == nil {
		return nil, NewAPIErrorWithID(nil, "read", connectorResourceName, connectorID)
	}

	rn := &remoteNetwork{
		ID:   r.Connector.RemoteNetwork.ID,
		Name: r.Connector.RemoteNetwork.Name,
	}

	connector := Connector{
		ID:            r.Connector.ID,
		Name:          r.Connector.Name,
		RemoteNetwork: rn,
	}

	return &connector, nil
}

type deleteConnectorQuery struct {
	ConnectorDelete *OkError `graphql:"connectorDelete(id: $id)" json:"connectorDelete"`
}

func (client *Client) deleteConnector(connectorID graphql.ID) error {
	if checkEmptyID(connectorID) {
		return NewAPIErrorWithID(ErrGraphqlIDIsEmpty, "delete", remoteNetworkResourceName, "connectorID")
	}

	variables := map[string]interface{}{
		"id": connectorID,
	}

	r := deleteConnectorQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	if !r.ConnectorDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(r.ConnectorDelete.Error), "delete", connectorResourceName, connectorID)
	}

	return nil
}
