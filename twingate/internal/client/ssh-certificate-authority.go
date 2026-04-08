package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
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

func (client *Client) ReadSSHCertificateAuthorities(ctx context.Context) ([]*model.CertificateAuthority, error) {
	opr := resourceSSHCertificateAuthority.read().withCustomName("readSSHCertificateAuthorities")

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
			return edge.Node.Type == "SSHCertificateAuthority"
		},
		func(edge *query.CertificateAuthorityEdge) *model.CertificateAuthority {
			return edge.Node.SSHCertificateAuthority.ToModel()
		}), nil
}

func (client *Client) readCertificateAuthoritiesAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.CertificateAuthorityEdge], error) {
	opr := resourceSSHCertificateAuthority.read().withCustomName("readCertificateAuthoritiesAfter")

	variables[query.CursorCertificateAuthorities] = cursor

	response := query.ReadCertificateAuthorities{}
	if err := client.query(ctx, &response, variables, opr, attr{}); err != nil {
		return nil, err
	}

	//nolint:staticcheck
	return &response.CertificateAuthorities.PaginatedResource, nil
}

func (client *Client) DeleteSSHCertificateAuthority(ctx context.Context, certificateAuthorityID string) error {
	opr := resourceSSHCertificateAuthority.delete()

	if certificateAuthorityID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteSSHCertificateAuthority{}

	return client.mutate(ctx, &response, newVars(gqlID(certificateAuthorityID)), opr, attr{id: certificateAuthorityID})
}
