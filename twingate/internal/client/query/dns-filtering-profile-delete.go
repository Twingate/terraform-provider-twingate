package query

type DeleteDNSFilteringProfile struct {
	OkError `graphql:"dnsFilteringProfileDelete(id: $id)"`
}

func (q DeleteDNSFilteringProfile) IsEmpty() bool {
	return false
}
