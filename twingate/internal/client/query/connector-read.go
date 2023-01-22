package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadConnector struct {
	Connector *gqlConnector `graphql:"connector(id: $id)"`
}

type gqlConnector struct {
	IDName
	RemoteNetwork struct {
		ID graphql.ID
	}
}

func (q ReadConnector) ToModel() *model.Connector {
	if q.Connector == nil {
		return nil
	}

	return q.Connector.ToModel()
}

func (c gqlConnector) ToModel() *model.Connector {
	return &model.Connector{
		ID:        c.StringID(),
		Name:      c.StringName(),
		NetworkID: idToString(c.RemoteNetwork.ID),
	}
}
