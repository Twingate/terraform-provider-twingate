package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
)

type ReadResource struct {
	Resource *gqlResource `graphql:"resource(id: $id)"`
}

type gqlResource struct {
	ResourceNode
	Groups Groups
}

type ResourceNode struct {
	IDName
	Address struct {
		Value graphql.String
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	Protocols *Protocols
	IsActive  graphql.Boolean
}

type Protocols struct {
	UDP       *Protocol       `json:"udp"`
	TCP       *Protocol       `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

type Protocol struct {
	Ports  []*PortRange   `json:"ports"`
	Policy graphql.String `json:"policy"`
}

type PortRange struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
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
