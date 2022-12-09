package query

type CreateResource struct {
	ResourceEntityResponse `graphql:"resourceCreate(name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

type ResourceEntityResponse struct {
	Entity *gqlResource
	OkError
}
