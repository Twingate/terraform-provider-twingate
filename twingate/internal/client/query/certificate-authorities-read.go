package query

const CursorCertificateAuthorities = "certificateAuthoritiesEndCursor"

type ReadCertificateAuthorities struct {
	CertificateAuthorities `graphql:"certificateAuthorities(after: $certificateAuthoritiesEndCursor, first: $pageLimit)"`
}

func (q ReadCertificateAuthorities) IsEmpty() bool {
	return len(q.Edges) == 0
}

type CertificateAuthorities struct {
	PaginatedResource[*CertificateAuthorityEdge]
}

type CertificateAuthorityEdge struct {
	Node *certificateAuthorityListNode
}

type certificateAuthorityListNode struct {
	Type                     string               `graphql:"__typename"`
	X509CertificateAuthority certificateAuthority `graphql:"... on X509CertificateAuthority"`
	SSHCertificateAuthority  certificateAuthority `graphql:"... on SSHCertificateAuthority"`
}
