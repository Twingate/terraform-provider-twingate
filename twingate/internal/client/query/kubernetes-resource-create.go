package query

import "github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"

//nolint:lll
type CreateKubernetesResource struct {
	KubernetesResourceEntityResponse `graphql:"kubernetesResourceCreate(name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId, isVisible: $isVisible, alias: $alias, securityPolicyId: $securityPolicyId, tags: $tags, accessPolicy: $accessPolicy, approvalMode: $approvalMode)"`
}

func (q CreateKubernetesResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateKubernetesResource) ToModel() *model.KubernetesResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

type KubernetesResourceEntityResponse struct {
	Entity *gqlKubernetesResource
	OkError
}

type gqlKubernetesResource struct {
	gqlPAMResource
}

func (g gqlKubernetesResource) ToModel() *model.KubernetesResource {
	return &model.KubernetesResource{PAMResource: g.gqlPAMResource.ToModel()}
}
