package query

type RemoveResourceAccess struct {
	OkError `graphql:"resourceAccessRemove(resourceId: $id, principalIds: $principalIds)"`
}

func (q RemoveResourceAccess) IsEmpty() bool {
	return false
}
