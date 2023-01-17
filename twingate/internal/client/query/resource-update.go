package query

type UpdateResource struct {
	ResourceEntityResponse `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols, isVisible: $isVisible, isBrowserShortcutEnabled: $isBrowserShortcutEnabled)"`
}

type UpdateResourceActiveState struct {
	OkError `graphql:"resourceUpdate(id: $id, isActive: $isActive)"`
}
