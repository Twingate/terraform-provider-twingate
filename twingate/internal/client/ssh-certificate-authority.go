//nolint:dupl
package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

func (client *Client) CreateSSHCertificateAuthority(ctx context.Context, name, publicKey string) (*model.CertificateAuthority, error) {
	opr := resourceSSHCertificateAuthority.create()

	if name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	if publicKey == "" {
		return nil, opr.apiError(ErrGraphqlPublicKeyIsEmpty)
	}

	variables := newVars(
		gqlVar(name, "name"),
		gqlVar(publicKey, "publicKey"),
	)

	response := query.CreateSSHCertificateAuthority{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: name}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadSSHCertificateAuthority(ctx context.Context, certificateAuthorityID string) (*model.CertificateAuthority, error) {
	opr := resourceSSHCertificateAuthority.read()

	if certificateAuthorityID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(certificateAuthorityID))
	response := query.ReadSSHCertificateAuthority{}

	if err := client.query(ctx, &response, variables, opr, attr{id: certificateAuthorityID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteSSHCertificateAuthority(ctx context.Context, certificateAuthorityID string) error {
	opr := resourceSSHCertificateAuthority.delete()

	if certificateAuthorityID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteSSHCertificateAuthority{}

	return client.mutate(ctx, &response, newVars(gqlID(certificateAuthorityID)), opr, attr{id: certificateAuthorityID})
}
