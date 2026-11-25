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
	gqlBaseResource

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
	pam := gqlPAMResource{gqlBaseResource: n.gqlBaseResource, Gateway: n.WebAppResourceFragment.Gateway}.ToModel()

	if err := setAccess(&pam.BaseResource, n.Access); err != nil {
		return nil, err
	}

	return &model.WebAppResource{
		PAMResource:           pam,
		Upstream:              model.WebAppUpstream{Port: n.WebAppResourceFragment.Upstream.Port, TLSMode: n.WebAppResourceFragment.Upstream.TLSMode},
		Downstream:            model.WebAppDownstream{Port: n.WebAppResourceFragment.Downstream.Port, TLSMode: n.WebAppResourceFragment.Downstream.TLSMode},
		RequestHeaderRewrites: headerRewritesToModel(n.WebAppResourceFragment.RequestHeaderRewrites),
	}, nil
}
