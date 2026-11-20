package query

import (
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadWebAppResource struct {
	Resource *gqlWebAppResourceNode `graphql:"resource(id: $id)"`
}

func (q ReadWebAppResource) IsEmpty() bool {
	return q.Resource == nil
}

func (q ReadWebAppResource) ToModel() (*model.WebAppResource, error) {
	if q.Resource == nil {
		return nil, nil //nolint:nilnil
	}

	return q.Resource.ToModel()
}

type gqlWebAppResourceNode struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	IsVisible              bool
	Alias                  string
	SecurityPolicy         *gqlSecurityPolicy
	Tags                   []Tag
	ApprovalMode           string
	AccessPolicy           *AccessPolicy
	Access                 Access `graphql:"access(after: $accessEndCursor, first: $pageLimit)"`
	WebAppResourceFragment struct {
		Gateway struct {
			ID graphql.ID
		}
		Upstream              WebAppUpstream
		Downstream            WebAppDownstream
		RequestHeaderRewrites []KeyValuePair
	} `graphql:"... on WebAppResource"`
}

func (n gqlWebAppResourceNode) ToModel() (*model.WebAppResource, error) {
	res := &model.WebAppResource{
		ID:                    string(n.ID),
		Name:                  n.Name,
		Address:               n.Address.Value,
		GatewayID:             string(n.WebAppResourceFragment.Gateway.ID),
		RemoteNetworkID:       string(n.RemoteNetwork.ID),
		IsVisible:             &n.IsVisible,
		Alias:                 optionalString(n.Alias),
		SecurityPolicyID:      securityPolicyID(n.SecurityPolicy),
		Tags:                  tagsToModel(n.Tags),
		Upstream:              model.WebAppUpstream{Port: n.WebAppResourceFragment.Upstream.Port, TLSMode: n.WebAppResourceFragment.Upstream.TLSMode},
		Downstream:            model.WebAppDownstream{Port: n.WebAppResourceFragment.Downstream.Port, TLSMode: n.WebAppResourceFragment.Downstream.TLSMode},
		RequestHeaderRewrites: headerRewritesToModel(n.WebAppResourceFragment.RequestHeaderRewrites),
		AccessPolicy:          accessPolicyToModel(n.AccessPolicy, &n.ApprovalMode),
	}

	var err error

	res.GroupsAccess, res.ServiceAccounts, err = accessToModel(n.Access.Edges)
	if err != nil {
		return nil, err
	}

	return res, nil
}
