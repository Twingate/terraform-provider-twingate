package query

type ReadSyncToS3OidcURL struct {
	OidcURL string `graphql:"eventsSyncOidcProviderUrl"`
}

func (q ReadSyncToS3OidcURL) IsEmpty() bool {
	// empty value is also valid
	return false
}
