package query

type CreateResource struct {
	ResourceEntityResponse `graphql:"resourceCreate(name: $name, address: $address, remoteNetworkId: $remoteNetworkId, protocols: $protocols, isVisible: $isVisible, isBrowserShortcutEnabled: $isBrowserShortcutEnabled, alias: $alias, securityPolicyId: $securityPolicyId, dlpPolicyId: $dlpPolicyId)"`
}

func (q CreateResource) IsEmpty() bool {
	return q.Entity == nil
}

type ResourceEntityResponse struct {
	Entity *gqlResource
	OkError
}
