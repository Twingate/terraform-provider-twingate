package query

type AddResourceGroups struct {
	OkError `graphql:"resourceUpdate(id: $id, addedGroupIds: $groupIds)"`
}

func (q AddResourceGroups) IsEmpty() bool {
	return false
}
