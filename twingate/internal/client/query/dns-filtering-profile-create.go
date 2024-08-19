package query

type CreateDNSFilteringProfile struct {
	DNSFilteringProfileEntityResponse `graphql:"dnsFilteringProfileCreate(name: $name)"`
}

type DNSFilteringProfileEntityResponse struct {
	Entity *gqlDNSFilteringProfile
	OkError
}

func (r *DNSFilteringProfileEntityResponse) IsEmpty() bool {
	return r == nil || r.Entity == nil
}
