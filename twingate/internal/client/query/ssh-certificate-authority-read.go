package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type ReadSSHCertificateAuthority struct {
	CertificateAuthority *certificateAuthorityNode `graphql:"certificateAuthority(id: $id)"`
}

func (q ReadSSHCertificateAuthority) IsEmpty() bool {
	return q.CertificateAuthority == nil
}

func (q ReadSSHCertificateAuthority) ToModel() *model.CertificateAuthority {
	if q.CertificateAuthority == nil || q.CertificateAuthority.Type != "SSHCertificateAuthority" {
		return nil
	}

	return q.CertificateAuthority.SSHCertificateAuthority.ToModel()
}
