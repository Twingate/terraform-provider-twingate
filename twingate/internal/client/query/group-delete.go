package query

type DeleteGroup struct {
	OkError `graphql:"groupDelete(id: $id)" json:"groupDelete"`
}
