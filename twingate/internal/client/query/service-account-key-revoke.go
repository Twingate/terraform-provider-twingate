package query

type RevokeServiceAccountKey struct {
	OkError `graphql:"serviceAccountKeyRevoke(id: $id)"`
}

func (q RevokeServiceAccountKey) IsEmpty() bool {
	return false
}
