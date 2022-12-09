package query

type UpdateGroup struct {
	GroupEntityResponse `graphql:"groupUpdate(id: $id, name: $name)"`
}
