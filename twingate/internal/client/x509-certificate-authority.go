package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
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

func (client *Client) ReadX509CertificateAuthorities(ctx context.Context) ([]*model.CertificateAuthority, error) {
	opr := resourceX509CertificateAuthority.read().withCustomName("readX509CertificateAuthorities")

	variables := newVars(
		cursor(query.CursorCertificateAuthorities),
		pageLimit(client.pageLimit),
	)

	response := query.ReadCertificateAuthorities{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil && !errors.Is(err, ErrGraphqlResultIsEmpty) {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readCertificateAuthoritiesAfter, variables); err != nil {
		return nil, err //nolint
	}

	return utils.FilterMap(response.Edges,
		func(edge *query.CertificateAuthorityEdge) bool {
			return edge.Node.Type == "X509CertificateAuthority"
		},
		func(edge *query.CertificateAuthorityEdge) *model.CertificateAuthority {
			return edge.Node.X509CertificateAuthority.ToModel()
		}), nil
}

func (client *Client) DeleteX509CertificateAuthority(ctx context.Context, certificateID string) error {
	opr := resourceX509CertificateAuthority.delete()

	if certificateID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteX509CertificateAuthority{}

	return client.mutate(ctx, &response, newVars(gqlID(certificateID)), opr, attr{id: certificateID})
}
