//nolint:dupl
package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

func (client *Client) CreateX509CertificateAuthority(ctx context.Context, name, certificate string) (*model.CertificateAuthority, error) {
	opr := resourceX509CertificateAuthority.create()

	if name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	if certificate == "" {
		return nil, opr.apiError(ErrGraphqlCertificateIsEmpty)
	}

	variables := newVars(
		gqlVar(name, "name"),
		gqlVar(certificate, "certificate"),
	)

	response := query.CreateX509CertificateAuthority{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: name}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadX509CertificateAuthority(ctx context.Context, certificateID string) (*model.CertificateAuthority, error) {
	opr := resourceX509CertificateAuthority.read()

	if certificateID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(certificateID))
	response := query.ReadX509CertificateAuthority{}

	if err := client.query(ctx, &response, variables, opr, attr{id: certificateID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteX509CertificateAuthority(ctx context.Context, certificateID string) error {
	opr := resourceX509CertificateAuthority.delete()

	if certificateID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteX509CertificateAuthority{}

	return client.mutate(ctx, &response, newVars(gqlID(certificateID)), opr, attr{id: certificateID})
}
