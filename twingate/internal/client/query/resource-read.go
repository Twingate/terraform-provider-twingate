package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

type ReadResource struct {
	Resource *gqlResource `graphql:"resource(id: $id)"`
}

func (q ReadResource) IsEmpty() bool {
	return q.Resource == nil
}

type gqlResource struct {
	ResourceNode
	Groups Groups `graphql:"groups(after: $groupsEndCursor, first: $pageLimit)"`
}

type ResourceNode struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	Protocols                *Protocols
	IsActive                 bool
	IsVisible                bool
	IsBrowserShortcutEnabled bool
	Alias                    string
}

type Protocols struct {
	UDP       *Protocol `json:"udp"`
	TCP       *Protocol `json:"tcp"`
	AllowIcmp bool      `json:"allowIcmp"`
}

type Protocol struct {
	Ports  []*PortRange `json:"ports"`
	Policy string       `json:"policy"`
}

type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func (r gqlResource) ToModel() *model.Resource {
	resource := r.ResourceNode.ToModel()
	resource.Groups = utils.Map[*GroupEdge, string](r.Groups.Edges, func(edge *GroupEdge) string {
		return string(edge.Node.ID)
	})

	return resource
}

func (r ResourceNode) ToModel() *model.Resource {
	return &model.Resource{
		ID:                       string(r.ID),
		Name:                     r.Name,
		Address:                  r.Address.Value,
		RemoteNetworkID:          string(r.RemoteNetwork.ID),
		Protocols:                protocolsToModel(r.Protocols),
		IsActive:                 r.IsActive,
		IsVisible:                &r.IsVisible,
		IsBrowserShortcutEnabled: &r.IsBrowserShortcutEnabled,
		Alias:                    optionalString(r.Alias),
	}
}

func protocolsToModel(protocols *Protocols) *model.Protocols {
	if protocols == nil {
		return nil
	}

	return &model.Protocols{
		UDP:       protocolToModel(protocols.UDP),
		TCP:       protocolToModel(protocols.TCP),
		AllowIcmp: protocols.AllowIcmp,
	}
}

func protocolToModel(protocol *Protocol) *model.Protocol {
	if protocol == nil {
		return nil
	}

	return &model.Protocol{
		Ports:  portsRangeToModel(protocol.Ports),
		Policy: protocol.Policy,
	}
}

func portsRangeToModel(ports []*PortRange) []*model.PortRange {
	return utils.Map[*PortRange, *model.PortRange](ports, func(port *PortRange) *model.PortRange {
		if port == nil {
			return nil
		}

		return &model.PortRange{
			Start: port.Start,
			End:   port.End,
		}
	})
}

func optionalString(str string) *string {
	if str == "" {
		return nil
	}

	return &str
}
