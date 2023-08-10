package query

type AddResourceAccess struct {
	OkError `graphql:"resourceAccessAdd(resourceId: $id, access: $access)"`
}

func (q AddResourceAccess) IsEmpty() bool {
	return false
}
