package query

type SetResourceAccess struct {
	OkError `graphql:"resourceAccessSet(resourceId: $id, access: $access)"`
}

func (q SetResourceAccess) IsEmpty() bool {
	return false
}
