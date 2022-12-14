package query

type DeleteServiceAccount struct {
	OkError `graphql:"serviceAccountDelete(id: $id)"`
}
