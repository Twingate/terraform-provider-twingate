package query

type UpdateGroup struct {
	GroupEntityResponse `graphql:"groupUpdate(id: $id, name: $name, addedUserIds: $addedUserIds, securityPolicyId: $securityPolicyId)"`
}

type UpdateGroupRemoveUsers struct {
	GroupEntityResponse `graphql:"groupUpdate(id: $id, removedUserIds: $removedUserIds)"`
}

func (q UpdateGroupRemoveUsers) IsEmpty() bool {
	return q.Entity == nil
}

func (q UpdateGroup) IsEmpty() bool {
	return q.Entity == nil
}
