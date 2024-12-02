package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
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
	Hostname                      string
	State                         string
	Version                       string
	PublicIP                      string   `graphql:"publicIP"`
	PrivateIPs                    []string `graphql:"privateIPs"`
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
		State:                c.State,
		Hostname:             c.Hostname,
		Version:              c.Version,
		PublicIP:             c.PublicIP,
		PrivateIPs:           c.PrivateIPs,
	}
}
