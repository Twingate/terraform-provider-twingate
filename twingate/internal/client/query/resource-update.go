package query

type UpdateResource struct {
	ResourceEntityResponse `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, addedGroupIds: $groupIds, protocols: $protocols)"`
}

type UpdateResourceActiveState struct {
	OkError `graphql:"resourceUpdate(id: $id, isActive: $isActive)"`
}

type UpdateResourceGroups struct {
	ResourceEntityResponse `graphql:"resourceUpdate(id: $id, groupIds: $groupIds)"`
}
