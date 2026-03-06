package query

type DeleteGateway struct {
	OkError `graphql:"gatewayDelete(id: $id)"`
}

func (q DeleteGateway) IsEmpty() bool {
	return false
}
