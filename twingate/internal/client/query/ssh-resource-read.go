package query

import (
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadSSHResource struct {
	Resource *gqlSSHResourceNode `graphql:"resource(id: $id)"`
}

func (q ReadSSHResource) IsEmpty() bool {
	return q.Resource == nil
}

func (q ReadSSHResource) ToModel() (*model.SSHResource, error) {
	if q.Resource == nil {
		return nil, nil //nolint:nilnil
	}

	return q.Resource.ToModel()
}

type gqlSSHResourceNode struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	IsVisible           bool
	Alias               string
	SecurityPolicy      *gqlSecurityPolicy
	Tags                []Tag
	ApprovalMode        string
	AccessPolicy        *AccessPolicy
	Access              Access `graphql:"access(after: $accessEndCursor, first: $pageLimit)"`
	SSHResourceFragment struct {
		Gateway struct {
			ID graphql.ID
		}
		Downstream SSHDownstream
		Upstream   SSHUpstream
	} `graphql:"... on SSHResource"`
}

func (n gqlSSHResourceNode) ToModel() (*model.SSHResource, error) {
	res := &model.SSHResource{
		ID:               string(n.ID),
		Name:             n.Name,
		Address:          n.Address.Value,
		GatewayID:        string(n.SSHResourceFragment.Gateway.ID),
		RemoteNetworkID:  string(n.RemoteNetwork.ID),
		IsVisible:        &n.IsVisible,
		Alias:            optionalString(n.Alias),
		SecurityPolicyID: securityPolicyID(n.SecurityPolicy),
		Tags:             tagsToModel(n.Tags),
		Downstream:       &model.SSHDownstream{Port: n.SSHResourceFragment.Downstream.Port},
		Upstream:         &model.SSHUpstream{Port: n.SSHResourceFragment.Upstream.Port},
		AccessPolicy:     accessPolicyToModel(n.AccessPolicy, &n.ApprovalMode),
	}

	groups, serviceAccounts, err := n.Access.ToModel()
	if err != nil {
		return nil, err
	}

	res.GroupsAccess = groups
	res.ServiceAccounts = serviceAccounts

	return res, nil
}
