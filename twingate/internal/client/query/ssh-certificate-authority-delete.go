package query

type DeleteSSHCertificateAuthority struct {
	OkError `graphql:"sshCertificateAuthorityDelete(id: $id)"`
}

func (q DeleteSSHCertificateAuthority) IsEmpty() bool {
	return false
}
