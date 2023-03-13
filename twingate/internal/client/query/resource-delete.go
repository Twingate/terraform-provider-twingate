package query

type DeleteResource struct {
	OkError `graphql:"resourceDelete(id: $id)"`
}

func (q DeleteResource) IsEmpty() bool {
	return false
}
