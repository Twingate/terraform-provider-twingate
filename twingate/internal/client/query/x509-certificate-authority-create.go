package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type CreateX509CertificateAuthority struct {
	X509CertificateAuthorityEntityResponse `graphql:"x509CertificateAuthorityCreate(name: $name, certificate: $certificate)"`
}

type X509CertificateAuthorityEntityResponse struct {
	Entity *certificateAuthority
	OkError
}

func (q CreateX509CertificateAuthority) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateX509CertificateAuthority) ToModel() *model.CertificateAuthority {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
