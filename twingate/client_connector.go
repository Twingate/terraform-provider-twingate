package twingate

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Connector struct {
	ID              string
	RemoteNetwork   *remoteNetwork
	Name            string
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

func (client *Client) createConnector(remoteNetworkID string) (*Connector, error) {
	variables := map[string]interface{}{
		"remoteNetworkId": graphql.ID(remoteNetworkID),
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
		ID:   r.ConnectorCreate.Entity.ID.(string),
		Name: string(r.ConnectorCreate.Entity.Name),
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
	Connector struct {
		IDName
		RemoteNetwork IDName
	} `graphql:"connector(id: $id)"`
}

func (client *Client) readConnector(connectorID string) (*Connector, error) {
	variables := map[string]interface{}{
		"id": graphql.ID(connectorID),
	}

	r := readConnectorQuery{}

	err := client.GraphqlClient.Query(context.Background(), &r, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	// if r.Connector == nil {
	// 	return nil, NewAPIErrorWithID(nil, "read", connectorResourceName, connectorID)
	// }

	rn := &remoteNetwork{
		ID:   r.Connector.RemoteNetwork.ID.(string),
		Name: string(r.Connector.RemoteNetwork.Name),
	}

	connector := Connector{
		ID:            r.Connector.ID.(string),
		Name:          string(r.Connector.Name),
		RemoteNetwork: rn,
	}

	return &connector, nil
}

type deleteConnectorQuery struct {
	ConnectorDelete *OkError `graphql:"connectorDelete(id: $id)"`
}

func (client *Client) deleteConnector(connectorID string) error {
	variables := map[string]interface{}{
		"id": graphql.ID(connectorID),
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
