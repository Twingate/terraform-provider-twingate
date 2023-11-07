package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

const (
	AccessGroup          = "Group"
	AccessServiceAccount = "ServiceAccount"
)

const CursorAccess = "accessEndCursor"

type ReadResource struct {
	Resource *gqlResource `graphql:"resource(id: $id)"`
}

func (q ReadResource) IsEmpty() bool {
	return q.Resource == nil
}

type gqlResource struct {
	ResourceNode
	Access Access `graphql:"access(after: $accessEndCursor, first: $pageLimit)"`
}

type Access struct {
	PaginatedResource[*AccessEdge]
}

type AccessEdge struct {
	Node           Principal
	SecurityPolicy *gqlSecurityPolicy
}

type Principal struct {
	Type string `graphql:"__typename"`
	Node `graphql:"... on Node"`
}

type Node struct {
	ID graphql.ID `json:"id"`
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
	SecurityPolicy           *gqlSecurityPolicy
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

	var (
		access          []model.ResourceAccess
		serviceAccounts []string
	)

	for _, edge := range r.Access.Edges {
		switch edge.Node.Type {
		case AccessGroup:
			resAccess := model.ResourceAccess{
				GroupID: optionalString(string(edge.Node.ID)),
			}

			if edge.SecurityPolicy != nil {
				resAccess.SecurityPolicyID = optionalString(string(edge.SecurityPolicy.ID))
			}

			access = append(access, resAccess)

		case AccessServiceAccount:
			serviceAccounts = append(serviceAccounts, string(edge.Node.ID))
		}
	}

	if len(serviceAccounts) > 0 {
		access = append(access, model.ResourceAccess{ServiceAccountIDs: serviceAccounts})
	}

	resource.Access = access

	return resource
}

func (r ResourceNode) ToModel() *model.Resource {
	var securityPolicyID string
	if r.SecurityPolicy != nil {
		securityPolicyID = string(r.SecurityPolicy.ID)
	}

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
		SecurityPolicyID:         optionalString(securityPolicyID),
	}
}

func protocolsToModel(protocols *Protocols) *model.Protocols {
	if protocols == nil {
		return model.DefaultProtocols()
	}

	return &model.Protocols{
		UDP:       protocolToModel(protocols.UDP),
		TCP:       protocolToModel(protocols.TCP),
		AllowIcmp: protocols.AllowIcmp,
	}
}

func protocolToModel(protocol *Protocol) *model.Protocol {
	if protocol == nil {
		return model.DefaultProtocol()
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
