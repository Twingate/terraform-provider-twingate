package transport

import (
	"fmt"

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

	return fmt.Sprintf("%v", id)
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

func (t gqlConnectorTokens) ToModel() *model.ConnectorTokens {
	return &model.ConnectorTokens{
		AccessToken:  string(t.AccessToken),
		RefreshToken: string(t.RefreshToken),
	}
}

func (q generateConnectorTokensQuery) ToModel() *model.ConnectorTokens {
	return q.ConnectorGenerateTokens.ConnectorTokens.ToModel()
}

func (q createGroupQuery) ToModel() *model.Group {
	return &model.Group{
		ID:   q.GroupCreate.Entity.StringID(),
		Name: q.GroupCreate.Entity.StringName(),
	}
}

func (g gqlGroup) ToModel() *model.Group {
	return &model.Group{
		ID:       g.StringID(),
		Name:     g.StringName(),
		Type:     string(g.Type),
		IsActive: bool(g.IsActive),
	}
}

func (gg gqlGroups) ToModel() []*model.Group {
	groups := make([]*model.Group, 0, len(gg.Edges))

	for _, g := range gg.Edges {
		if g == nil || g.Node == nil {
			continue
		}

		groups = append(groups, g.Node.ToModel())
	}

	if cap(groups) > len(groups) {
		groups = groups[:len(groups):len(groups)]
	}

	return groups
}

func (q readGroupQuery) ToModel() *model.Group {
	if q.Group == nil {
		return nil
	}

	return q.Group.ToModel()
}

func (q readGroupsQuery) ToModel() []*model.Group {
	return q.Groups.ToModel()
}

func (q readGroupsByNameQuery) ToModel() []*model.Group {
	return q.Groups.ToModel()
}

func (n gqlRemoteNetwork) ToModel() *model.RemoteNetwork {
	return &model.RemoteNetwork{
		ID:   idToString(n.ID),
		Name: string(n.Name),
	}
}

func (nn gqlRemoteNetworks) ToModel() []*model.RemoteNetwork {
	networks := make([]*model.RemoteNetwork, 0, len(nn.Edges))

	for _, network := range nn.Edges {
		if network == nil {
			continue
		}

		networks = append(networks, network.Node.ToModel())
	}

	if cap(networks) > len(networks) {
		networks = networks[:len(networks):len(networks)]
	}

	return networks
}

func (q createRemoteNetworkQuery) ToModel() *model.RemoteNetwork {
	return q.RemoteNetworkCreate.Entity.ToModel()
}

func (q readRemoteNetworksQuery) ToModel() []*model.RemoteNetwork {
	return q.RemoteNetworks.ToModel()
}
