package query

type DeleteServiceAccountKey struct {
	OkError `graphql:"serviceAccountKeyDelete(id: $id)"`
}

func (q DeleteServiceAccountKey) IsEmpty() bool {
	return false
}
