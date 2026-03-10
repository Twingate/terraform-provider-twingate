//nolint:dupl
package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
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
	Protocols           *Protocols
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
		Protocols:        protocolsToModel(n.Protocols),
		AccessPolicy:     accessPolicyToModel(n.AccessPolicy, &n.ApprovalMode),
	}

	for _, access := range n.Access.Edges {
		if access.Node.Type != AccessGroup {
			continue
		}

		groupID := string(access.Node.Group.ID)
		if groupID == "" {
			return nil, ErrMissingAccessGroupID
		}

		var secPolicyID *string
		if access.SecurityPolicy != nil {
			secPolicyID = optionalString(string(access.SecurityPolicy.ID))
		}

		res.GroupsAccess = append(res.GroupsAccess, model.AccessGroup{
			GroupID:          groupID,
			SecurityPolicyID: secPolicyID,
			AccessPolicy:     accessPolicyToModel(access.AccessPolicy, access.ApprovalMode),
		})
	}

	return res, nil
}
