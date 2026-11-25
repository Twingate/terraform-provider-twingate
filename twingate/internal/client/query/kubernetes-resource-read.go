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
	gqlBaseResource

	Access                     Access `graphql:"access(after: $accessEndCursor, first: $pageLimit)"`
	KubernetesResourceFragment struct {
		Gateway struct {
			ID graphql.ID
		}
	} `graphql:"... on KubernetesResource"`
}

func (n gqlKubernetesResourceNode) ToModel() (*model.KubernetesResource, error) {
	pam := gqlPAMResource{gqlBaseResource: n.gqlBaseResource, Gateway: n.KubernetesResourceFragment.Gateway}.ToModel()

	if err := setAccess(&pam.BaseResource, n.Access); err != nil {
		return nil, err
	}

	return &model.KubernetesResource{
		PAMResource: pam,
	}, nil
}
