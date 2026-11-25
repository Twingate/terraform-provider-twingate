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
	gqlBaseResource

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
	pam := gqlPAMResource{gqlBaseResource: n.gqlBaseResource, Gateway: n.SSHResourceFragment.Gateway}.ToModel()

	if err := setAccess(&pam.BaseResource, n.Access); err != nil {
		return nil, err
	}

	return &model.SSHResource{
		PAMResource: pam,
		Downstream:  &model.SSHDownstream{Port: n.SSHResourceFragment.Downstream.Port},
		Upstream:    &model.SSHUpstream{Port: n.SSHResourceFragment.Upstream.Port},
	}, nil
}
