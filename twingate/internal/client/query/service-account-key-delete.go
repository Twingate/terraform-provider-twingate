package query

type DeleteServiceAccountKey struct {
	OkError `graphql:"serviceAccountKeyDelete(id: $id)"`
}
