package query

type DeleteX509CertificateAuthority struct {
	OkError `graphql:"x509CertificateAuthorityDelete(id: $id)"`
}

func (q DeleteX509CertificateAuthority) IsEmpty() bool {
	return false
}
