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
	HasStatusNotificationsEnabled bool
}

func (q ReadConnector) IsEmpty() bool {
	return q.Connector == nil
}

func (q ReadConnector) ToModel() *model.Connector {
	if q.Connector == nil {
		return nil
	}

	return q.Connector.ToModel()
}

func (c gqlConnector) ToModel() *model.Connector {
	return &model.Connector{
		ID:                   string(c.ID),
		Name:                 c.Name,
		NetworkID:            string(c.RemoteNetwork.ID),
		StatusUpdatesEnabled: &c.HasStatusNotificationsEnabled,
	}
}
