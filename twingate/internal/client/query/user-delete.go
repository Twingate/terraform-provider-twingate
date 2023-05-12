package query

type DeleteUser struct {
	OkError `graphql:"userDelete(id: $id)"`
}

func (q DeleteUser) IsEmpty() bool {
	return false
}
