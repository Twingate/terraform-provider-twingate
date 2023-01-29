package query

type UpdateGroup struct {
	GroupEntityResponse `graphql:"groupUpdate(id: $id, name: $name, addedUserIds: $addedUserIds)"`
}

type UpdateGroupRemoveUsers struct {
	GroupEntityResponse `graphql:"groupUpdate(id: $id, removedUserIds: $removedUserIds)"`
}
