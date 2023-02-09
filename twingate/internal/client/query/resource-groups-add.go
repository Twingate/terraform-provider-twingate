package query

type AddResourceGroups struct {
	OkError `graphql:"resourceUpdate(id: $id, addedGroupIds: $groupIds)"`
}
