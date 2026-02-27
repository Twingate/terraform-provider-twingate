package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadX509CertificateAuthority struct {
	CertificateAuthority *certificateAuthorityNode `graphql:"certificateAuthority(id: $id)"`
}

func (q ReadX509CertificateAuthority) IsEmpty() bool {
	return q.CertificateAuthority == nil
}

func (q ReadX509CertificateAuthority) ToModel() *model.CertificateAuthority {
	if q.CertificateAuthority == nil || q.CertificateAuthority.Type != "X509CertificateAuthority" {
		return nil
	}

	return q.CertificateAuthority.X509CertificateAuthority.ToModel()
}

type certificateAuthorityNode struct {
	Type                     string               `graphql:"__typename"`
	X509CertificateAuthority certificateAuthority `graphql:"... on X509CertificateAuthority"`
	SSHCertificateAuthority  certificateAuthority `graphql:"... on SSHCertificateAuthority"`
}

type certificateAuthority struct {
	ID          graphql.ID
	Name        string
	Fingerprint string
}

func (g certificateAuthority) ToModel() *model.CertificateAuthority {
	return &model.CertificateAuthority{
		ID:          string(g.ID),
		Name:        g.Name,
		Fingerprint: g.Fingerprint,
	}
}
