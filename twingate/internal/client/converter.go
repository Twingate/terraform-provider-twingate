package client

import (
	"fmt"
	"time"

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
		UserIDs: utils.Map[*UserEdge, string](g.Users.Edges, func(edge *UserEdge) string {
			return edge.Node.ID.(string)
		}),
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

func (q createServiceAccountQuery) ToModel() *model.ServiceAccount {
	return &model.ServiceAccount{
		ID:   q.ServiceAccountCreate.Entity.StringID(),
		Name: q.ServiceAccountCreate.Entity.StringName(),
	}
}

func (q readServiceAccountQuery) ToModel() *model.ServiceAccount {
	if q.ServiceAccount == nil {
		return nil
	}

	return q.ServiceAccount.ToModel()
}

func (q gqlServiceAccount) ToModel() *model.ServiceAccount {
	return &model.ServiceAccount{
		ID:   q.StringID(),
		Name: q.StringName(),
	}
}

func (s *ServiceAccounts) ToModel() []*model.ServiceAccount {
	return utils.Map[*ServiceAccountEdge, *model.ServiceAccount](s.Edges, func(edge *ServiceAccountEdge) *model.ServiceAccount {
		return edge.Node.ToModel()
	})
}

func (q gqlServiceKey) ToModel() (*model.ServiceKey, error) {
	expirationTime, err := q.parseExpirationTime()
	if err != nil {
		return nil, err
	}

	return &model.ServiceKey{
		ID:             q.StringID(),
		Name:           q.StringName(),
		Service:        q.ServiceAccount.StringID(),
		ExpirationTime: expirationTime,
		Status:         string(q.Status),
	}, nil
}

func (q gqlServiceKey) parseExpirationTime() (int, error) {
	if q.ExpiresAt == "" {
		return 0, nil
	}

	expiresAt, err := time.Parse(time.RFC3339, string(q.ExpiresAt))
	if err != nil {
		return -1, fmt.Errorf("failed to parse expiration time `%s`: %w", q.ExpiresAt, err)
	}

	return getDaysTillExpiration(expiresAt), nil
}

func getDaysTillExpiration(expiresAt time.Time) int {
	const hoursInDay = 24

	return int(time.Until(expiresAt).Hours()/hoursInDay) + 1
}

func (q createServiceAccountKeyQuery) ToModel() (*model.ServiceKey, error) {
	return q.ServiceAccountKeyCreate.Entity.ToModel()
}

func (q readServiceAccountKeyQuery) ToModel() (*model.ServiceKey, error) {
	return q.ServiceAccountKey.ToModel()
}

func (q updateServiceAccountKeyQuery) ToModel() (*model.ServiceKey, error) {
	return q.ServiceAccountKeyUpdate.Entity.ToModel()
}

func (s *Services) ToModel() []*model.Service {
	return utils.Map[*ServiceEdge, *model.Service](s.Edges, func(edge *ServiceEdge) *model.Service {
		return edge.Node.ToModel()
	})
}

func (s *gqlService) ToModel() *model.Service {
	return &model.Service{
		ID:        s.StringID(),
		Name:      s.StringName(),
		Resources: s.Resources.listIDs(),
		Keys:      s.Keys.listIDs(),
	}
}

func (q gqlResourceIDs) listIDs() []string {
	return utils.Map[*gqlResourceIDEdge, string](q.Edges, func(edge *gqlResourceIDEdge) string {
		return edge.Node.ID.(string)
	})
}

func (q gqlKeyIDs) listIDs() []string {
	return utils.Map[*gqlKeyIDEdge, string](q.Edges, func(edge *gqlKeyIDEdge) string {
		return edge.Node.ID.(string)
	})
}
