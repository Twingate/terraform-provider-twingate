package twingate

import (
	"fmt"
)

type Connector struct {
	Id            string
	RemoteNetwork *RemoteNetwork
	Name          string
	AccessToken   string
	RefreshToken  string
}

func (client *Client) createConnector(remoteNetworkId *string) (*Connector, error) {
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
        `, *remoteNetworkId),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, err
	}

	connectorResult := mutationConnector.Path("data.connectorCreate")
	status := connectorResult.Path("ok").Data().(bool)

	if !status {
		errorString := connectorResult.Path("error").Data().(string)
		return nil, fmt.Errorf("Cant create connector under the network with name %s, Error:  %s", *remoteNetworkId, errorString)
	}

	connector := Connector{
		Id:   connectorResult.Path("entity.id").Data().(string),
		Name: connectorResult.Path("entity.name").Data().(string),
	}
	return &connector, nil
}

func (client *Client) readConnector(connectorId *string) (*Connector, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  connector(id: "%s") {
			id
			name
			RemoteNetwork {
				name
				id
			}
          }
		}
        `, *connectorId),
	}
	queryConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, err
	}

	connectorRead := queryConnector.Path("data.connector")

	if connectorRead.Data() == nil {
		return nil, fmt.Errorf("Unable to read connector %s information ", *connectorId)
	}

	connector := Connector{
		Id:   connectorRead.Path("id").Data().(string),
		Name: connectorRead.Path("name").Data().(string),
		RemoteNetwork: &RemoteNetwork{
			Id:   connectorRead.Path("RemoteNetwork.id").Data().(string),
			Name: connectorRead.Path("RemoteNetwork.name").Data().(string),
		},
	}

	return &connector, nil
}

func (client *Client) deleteConnector(connectorId *string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  connectorDelete(id: "%s"){
			ok
			error
		  }
		}
		`, *connectorId),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return err
	}
	connectorDelete := mutationConnector.Path("data.connectorDelete")
	status := connectorDelete.Path("ok").Data().(bool)
	if !status {
		return fmt.Errorf("Unable to delete network with Id %s, Error:  %s", *connectorId, connectorDelete.Path("error").Data().(string))
	}

	return nil
}
