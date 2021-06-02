package twingate

import (
	"fmt"
)

type Connector struct {
	ID              string
	RemoteNetwork   *RemoteNetwork
	Name            string
	ConnectorTokens *ConnectorTokens
}

type Connectors struct {
	ID   string
	Name string
}

const connectorResourceName = "connector"

func (client *Client) createConnector(remoteNetworkID string) (*Connector, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  connectorCreate(remoteNetworkId: "%s"){
				ok
				error
				entity{
				  id
                  name
				}
			  }
			}
        `, remoteNetworkID),
	}

	mutationConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	connectorResult := mutationConnector.Path("data.connectorCreate")

	status := connectorResult.Path("ok").Data().(bool)
	if !status {
		message := connectorResult.Path("error").Data().(string)

		return nil, NewAPIError(NewMutationError(message), "create", connectorResourceName)
	}

	connector := Connector{
		ID:   connectorResult.Path("entity.id").Data().(string),
		Name: connectorResult.Path("entity.name").Data().(string),
	}

	return &connector, nil
}

func (client *Client) readConnectors() (map[int]*Connectors, error) { //nolint
	query := map[string]string{
		"query": "{ connectors { edges { node { id name } } } }",
	}

	queryResource, err := client.doGraphqlRequest(query)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	var connectors = make(map[int]*Connectors)

	queryChildren := queryResource.Path("data.connectors.edges").Children()

	for i, elem := range queryChildren {
		nodeID := elem.Path("node.id").Data().(string)
		nodeName := elem.Path("node.name").Data().(string)
		c := &Connectors{ID: nodeID, Name: nodeName}
		connectors[i] = c
	}

	return connectors, nil
}

func (client *Client) readConnector(connectorID string) (*Connector, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  connector(id: "%s") {
			id
			name
			remoteNetwork {
				name
				id
			}
          }
		}
        `, connectorID),
	}

	queryConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	connectorRead := queryConnector.Path("data.connector")
	if connectorRead.Data() == nil {
		return nil, NewAPIErrorWithID(nil, "read", connectorResourceName, connectorID)
	}

	connector := Connector{
		ID:   connectorRead.Path("id").Data().(string),
		Name: connectorRead.Path("name").Data().(string),
		RemoteNetwork: &RemoteNetwork{
			ID:   connectorRead.Path("remoteNetwork.id").Data().(string),
			Name: connectorRead.Path("remoteNetwork.name").Data().(string),
		},
	}

	return &connector, nil
}

func (client *Client) deleteConnector(connectorID string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  connectorDelete(id: "%s"){
			ok
			error
		  }
		}
		`, connectorID),
	}

	mutationConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	connectorDelete := mutationConnector.Path("data.connectorDelete")

	status := connectorDelete.Path("ok").Data().(bool)
	if !status {
		message := connectorDelete.Path("error").Data().(string)

		return NewAPIErrorWithID(NewMutationError(message), "delete", connectorResourceName, connectorID)
	}

	return nil
}
