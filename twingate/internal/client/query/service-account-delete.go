package query

type DeleteServiceAccount struct {
	OkError `graphql:"serviceAccountDelete(id: $id)"`
}

func (q DeleteServiceAccount) IsEmpty() bool {
	return false
}
