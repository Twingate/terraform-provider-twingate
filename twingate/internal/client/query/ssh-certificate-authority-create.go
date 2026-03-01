package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type CreateSSHCertificateAuthority struct {
	SSHCertificateAuthorityEntityResponse `graphql:"sshCertificateAuthorityCreate(name: $name, publicKey: $publicKey)"`
}

type SSHCertificateAuthorityEntityResponse struct {
	Entity *certificateAuthority
	OkError
}

func (q CreateSSHCertificateAuthority) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateSSHCertificateAuthority) ToModel() *model.CertificateAuthority {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
