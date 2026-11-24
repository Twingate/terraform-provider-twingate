package query

import (
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadKubernetesResource struct {
	Resource *gqlKubernetesResourceNode `graphql:"resource(id: $id)"`
}

func (q ReadKubernetesResource) IsEmpty() bool {
	return q.Resource == nil
}

func (q ReadKubernetesResource) ToModel() (*model.KubernetesResource, error) {
	if q.Resource == nil {
		return nil, nil //nolint:nilnil
	}

	return q.Resource.ToModel()
}

type gqlKubernetesResourceNode struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	IsVisible                  bool
	Alias                      string
	SecurityPolicy             *gqlSecurityPolicy
	Tags                       []Tag
	ApprovalMode               string
	AccessPolicy               *AccessPolicy
	Access                     Access `graphql:"access(after: $accessEndCursor, first: $pageLimit)"`
	KubernetesResourceFragment struct {
		Gateway struct {
			ID graphql.ID
		}
	} `graphql:"... on KubernetesResource"`
}

func (n gqlKubernetesResourceNode) ToModel() (*model.KubernetesResource, error) {
	res := &model.KubernetesResource{
		ID:               string(n.ID),
		Name:             n.Name,
		Address:          n.Address.Value,
		GatewayID:        string(n.KubernetesResourceFragment.Gateway.ID),
		RemoteNetworkID:  string(n.RemoteNetwork.ID),
		IsVisible:        &n.IsVisible,
		Alias:            optionalString(n.Alias),
		SecurityPolicyID: securityPolicyID(n.SecurityPolicy),
		Tags:             tagsToModel(n.Tags),
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
