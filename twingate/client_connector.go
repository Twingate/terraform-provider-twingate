package twingate

import (
	"fmt"
)

type Connector struct {
	Id              string
	RemoteNetwork   *RemoteNetwork
	Name            string
	ConnectorTokens *ConnectorTokens
}

func (client *Client) createConnector(remoteNetworkId string) (*Connector, error) {
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
        `, remoteNetworkId),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIError(err, "create", "connector")
	}

	connectorResult := mutationConnector.Path("data.connectorCreate")
	status := connectorResult.Path("ok").Data().(bool)

	if !status {
		message := connectorResult.Path("error").Data().(string)

		return nil, NewAPIError(NewMutationError(message), "create", "connector")
	}

	connector := Connector{
		Id:   connectorResult.Path("entity.id").Data().(string),
		Name: connectorResult.Path("entity.name").Data().(string),
	}

	return &connector, nil
}

func (client *Client) readConnector(connectorId string) (*Connector, error) {
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
        `, connectorId),
	}
	queryConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIErrorWithId(err, "reed", "connector", connectorId)
	}

	connectorRead := queryConnector.Path("data.connector")

	if connectorRead.Data() == nil {
		return nil, NewAPIErrorWithId(nil, "reed", "connector", connectorId)
	}

	connector := Connector{
		Id:   connectorRead.Path("id").Data().(string),
		Name: connectorRead.Path("name").Data().(string),
		RemoteNetwork: &RemoteNetwork{
			Id:   connectorRead.Path("remoteNetwork.id").Data().(string),
			Name: connectorRead.Path("remoteNetwork.name").Data().(string),
		},
	}

	return &connector, nil
}

func (client *Client) deleteConnector(connectorId string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  connectorDelete(id: "%s"){
			ok
			error
		  }
		}
		`, connectorId),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return NewAPIErrorWithId(err, "delete", "connector", connectorId)
	}
	connectorDelete := mutationConnector.Path("data.connectorDelete")
	status := connectorDelete.Path("ok").Data().(bool)
	if !status {
		message := connectorDelete.Path("error").Data().(string)

		return NewAPIErrorWithId(NewMutationError(message), "delete", "connector", connectorId)
	}

	return nil
}
