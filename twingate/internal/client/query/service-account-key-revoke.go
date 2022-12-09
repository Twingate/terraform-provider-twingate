package query

type RevokeServiceAccountKey struct {
	OkError `graphql:"serviceAccountKeyRevoke(id: $id)"`
}
