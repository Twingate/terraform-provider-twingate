package query

type UpdateResource struct {
	ResourceEntityResponse `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, protocols: $protocols, isVisible: $isVisible, isBrowserShortcutEnabled: $isBrowserShortcutEnabled, alias: $alias, securityPolicyId: $securityPolicyId)"`
}

func (q UpdateResource) IsEmpty() bool {
	return q.Entity == nil
}

type UpdateResourceActiveState struct {
	OkError `graphql:"resourceUpdate(id: $id, isActive: $isActive)"`
}

func (q UpdateResourceActiveState) IsEmpty() bool {
	return false
}

type UpdateResourceRemoveGroups struct {
	ResourceEntityResponse `graphql:"resourceUpdate(id: $id, removedGroupIds: $removedGroupIds)"`
}

func (q UpdateResourceRemoveGroups) IsEmpty() bool {
	return q.Entity == nil
}
