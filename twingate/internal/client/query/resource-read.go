package query

import (
	"errors"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

var (
	ErrMissingAccessGroupID          = errors.New("access group ID is missing in response")
	ErrMissingAccessServiceAccountID = errors.New("access service account ID is missing in response")
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
	ApprovalMode   *string
	AccessPolicy   *AccessPolicy
}

type Principal struct {
	Type           string `graphql:"__typename"`
	Group          Node   `graphql:"... on Group"`
	ServiceAccount Node   `graphql:"... on ServiceAccount"`
}

type Node struct {
	ID graphql.ID `json:"id"`
}

type Tag struct {
	Key   string
	Value string
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
	Tags                     []Tag
	ApprovalMode             string
	AccessPolicy             *AccessPolicy
}

type AccessPolicy struct {
	DurationSeconds *int64
	Mode            AccessMode
}

type AccessMode string

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

func (r gqlResource) ToModel() (*model.Resource, error) {
	resource := r.ResourceNode.ToModel()

	for _, access := range r.Access.Edges {
		var securityPolicyID *string
		if access.SecurityPolicy != nil {
			securityPolicyID = optionalString(string(access.SecurityPolicy.ID))
		}

		switch access.Node.Type {
		case AccessGroup:
			groupID := string(access.Node.Group.ID)
			if groupID == "" {
				return nil, ErrMissingAccessGroupID
			}

			resource.GroupsAccess = append(resource.GroupsAccess, model.AccessGroup{
				GroupID:          groupID,
				SecurityPolicyID: securityPolicyID,
				AccessPolicy:     accessPolicyToModel(access.AccessPolicy, access.ApprovalMode),
			})
		case AccessServiceAccount:
			serviceAccountID := string(access.Node.ServiceAccount.ID)
			if serviceAccountID == "" {
				return nil, ErrMissingAccessServiceAccountID
			}

			resource.ServiceAccounts = append(resource.ServiceAccounts, serviceAccountID)
		}
	}

	return resource, nil
}

func (r ResourceNode) ToModel() *model.Resource {
	var securityPolicy string
	if r.SecurityPolicy != nil {
		securityPolicy = string(r.SecurityPolicy.ID)
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
		SecurityPolicyID:         optionalString(securityPolicy),
		Tags:                     tagsToModel(r.Tags),
		AccessPolicy:             accessPolicyToModel(r.AccessPolicy, &r.ApprovalMode),
	}
}

func accessPolicyToModel(accessPolicy *AccessPolicy, approvalMode *string) *model.AccessPolicy {
	if accessPolicy == nil {
		return nil
	}

	var duration *string

	if accessPolicy.DurationSeconds != nil {
		val := time.Duration(*accessPolicy.DurationSeconds) * time.Second
		raw := utils.FormatDurationWithDays(val)
		duration = &raw
	}

	mode := string(accessPolicy.Mode)

	return &model.AccessPolicy{
		Mode:         &mode,
		Duration:     duration,
		ApprovalMode: approvalMode,
	}
}

func tagsToModel(tags []Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	m := make(map[string]string, len(tags))
	for _, tag := range tags {
		m[tag.Key] = tag.Value
	}

	return m
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
