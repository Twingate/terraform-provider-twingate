package client

import (
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
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

	return q.Connectors.ToModel()
}

func (c *Connectors) ToModel() []*model.Connector {
	return utils.Map[*ConnectorEdge, *model.Connector](c.Edges, func(edge *ConnectorEdge) *model.Connector {
		return edge.Node.ToModel()
	})
}

func (r *Resources) ToModel() []*model.Resource {
	return utils.Map[*ResourceEdge, *model.Resource](r.Edges, func(edge *ResourceEdge) *model.Resource {
		return edge.Node.ToModel()
	})
}

func (r *RemoteNetworks) ToModel() []*model.RemoteNetwork {
	return utils.Map[*gqlRemoteNetworkEdge, *model.RemoteNetwork](r.Edges, func(edge *gqlRemoteNetworkEdge) *model.RemoteNetwork {
		return edge.Node.ToModel()
	})
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

func (q readGroupQuery) ToModel() *model.Group {
	if q.Group == nil {
		return nil
	}

	return q.Group.ToModel()
}

func (n gqlRemoteNetwork) ToModel() *model.RemoteNetwork {
	return &model.RemoteNetwork{
		ID:   idToString(n.ID),
		Name: string(n.Name),
	}
}

func (nn gqlRemoteNetworks) ToModel() []*model.RemoteNetwork {
	return utils.Map[*gqlRemoteNetworkEdge, *model.RemoteNetwork](nn.Edges, func(edge *gqlRemoteNetworkEdge) *model.RemoteNetwork {
		return edge.Node.ToModel()
	})
}

func (q createRemoteNetworkQuery) ToModel() *model.RemoteNetwork {
	return q.RemoteNetworkCreate.Entity.ToModel()
}

func (q readRemoteNetworksQuery) ToModel() []*model.RemoteNetwork {
	return q.RemoteNetworks.ToModel()
}

func (u gqlUser) ToModel() *model.User {
	return &model.User{
		ID:        idToString(u.ID),
		FirstName: string(u.FirstName),
		LastName:  string(u.LastName),
		Email:     string(u.Email),
		Role:      string(u.Role),
	}
}

func (u *Users) ToModel() []*model.User {
	return utils.Map[*UserEdge, *model.User](u.Edges, func(edge *UserEdge) *model.User {
		return edge.Node.ToModel()
	})
}

func (u *Groups) ToModel() []*model.Group {
	return utils.Map[*GroupEdge, *model.Group](u.Edges, func(edge *GroupEdge) *model.Group {
		return edge.Node.ToModel()
	})
}

func newProtocolsInput(protocols *model.Protocols) *ProtocolsInput {
	if protocols == nil {
		return nil
	}

	return &ProtocolsInput{
		UDP:       newProtocol(protocols.UDP),
		TCP:       newProtocol(protocols.TCP),
		AllowIcmp: graphql.Boolean(protocols.AllowIcmp),
	}
}

func newProtocol(protocol *model.Protocol) *ProtocolInput {
	if protocol == nil {
		return nil
	}

	return &ProtocolInput{
		Ports:  newPorts(protocol.Ports),
		Policy: graphql.String(protocol.Policy),
	}
}

func newPorts(ports []*model.PortRange) []*PortRangeInput {
	return utils.Map[*model.PortRange, *PortRangeInput](ports, func(port *model.PortRange) *PortRangeInput {
		return &PortRangeInput{
			Start: graphql.Int(port.Start),
			End:   graphql.Int(port.End),
		}
	})
}

func protocolsToModel(protocols *Protocols) *model.Protocols {
	if protocols == nil {
		return nil
	}

	return &model.Protocols{
		UDP:       protocolToModel(protocols.UDP),
		TCP:       protocolToModel(protocols.TCP),
		AllowIcmp: bool(protocols.AllowIcmp),
	}
}

func protocolToModel(protocol *Protocol) *model.Protocol {
	if protocol == nil {
		return nil
	}

	return &model.Protocol{
		Ports:  portsRangeToModel(protocol.Ports),
		Policy: string(protocol.Policy),
	}
}

func portsRangeToModel(ports []*PortRange) []*model.PortRange {
	return utils.Map[*PortRange, *model.PortRange](ports, func(port *PortRange) *model.PortRange {
		if port == nil {
			return nil
		}

		return &model.PortRange{
			Start: int32(port.Start),
			End:   int32(port.End),
		}
	})
}

func (r gqlResource) ToModel() *model.Resource {
	groups := make([]string, 0, len(r.Groups.Edges))
	for _, elem := range r.Groups.Edges {
		groups = append(groups, idToString(elem.Node.ID))
	}

	resource := r.ResourceNode.ToModel()
	resource.Groups = groups

	return resource
}

func (r ResourceNode) ToModel() *model.Resource {
	return &model.Resource{
		ID:              r.StringID(),
		Name:            r.StringName(),
		Address:         string(r.Address.Value),
		RemoteNetworkID: idToString(r.RemoteNetwork.ID),
		Protocols:       protocolsToModel(r.Protocols),
		IsActive:        bool(r.IsActive),
	}
}

func (q readResourcesByNameQuery) ToModel() []*model.Resource {
	return q.Resources.ToModel()
}
