package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type UpdateKubernetesResource struct {
	KubernetesResourceEntityResponse `graphql:"kubernetesResourceUpdate(id: $id, name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId)"`
}

func (q UpdateKubernetesResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q UpdateKubernetesResource) ToModel() *model.KubernetesResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
