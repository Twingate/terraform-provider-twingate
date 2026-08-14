package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

//nolint:lll
type UpdateWebAppResource struct {
	WebAppResourceEntityResponse `graphql:"webAppResourceUpdate(id: $id, name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId, isVisible: $isVisible, alias: $alias, securityPolicyId: $securityPolicyId, tags: $tags, upstream: $upstream, downstream: $downstream, requestHeaderRewrites: $requestHeaderRewrites, accessPolicy: $accessPolicy, approvalMode: $approvalMode)"`
}

func (q UpdateWebAppResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q UpdateWebAppResource) ToModel() *model.WebAppResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
