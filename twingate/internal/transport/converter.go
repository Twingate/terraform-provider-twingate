package transport

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

func (c gqlConnector) ToModel() *model.Connector {
	return &model.Connector{
		ID:        c.StringID(),
		Name:      c.StringName(),
		NetworkID: idToString(c.RemoteNetwork.ID),
	}
}

func idToString(id graphql.ID) string {
	if id == nil {
		return ""
	}

	return id.(string)
}

func (q readConnectorQuery) ToModel() *model.Connector {
	if q.Connector == nil {
		return nil
	}

	return q.Connector.ToModel()
}

func (q readConnectorsQuery) ToModel() []*model.Connector {
	if len(q.Connectors.Edges) == 0 {
		return nil
	}

	connectors := make([]*model.Connector, 0, len(q.Connectors.Edges))

	for _, elem := range q.Connectors.Edges {
		if elem == nil {
			continue
		}

		connectors = append(connectors, elem.Node.ToModel())
	}

	if cap(connectors) > len(connectors) {
		connectors = connectors[:len(connectors):len(connectors)]
	}

	return connectors
}

func (q createConnectorQuery) ToModel() *model.Connector {
	return q.ConnectorCreate.Entity.ToModel()
}
