package query

type DeleteResource struct {
	OkError `graphql:"resourceDelete(id: $id)"`
}
