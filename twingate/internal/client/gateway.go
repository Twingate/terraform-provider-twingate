package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

func (client *Client) CreateGateway(ctx context.Context, address, remoteNetworkID, x509CAID, sshCAID string) (*model.Gateway, error) {
	opr := resourceGateway.create()

	if address == "" {
		return nil, opr.apiError(ErrGraphqlAddressIsEmpty)
	}

	if remoteNetworkID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if x509CAID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlVar(address, "address"),
		gqlID(remoteNetworkID, "remoteNetworkId"),
		gqlID(x509CAID, "x509CAId"),
		gqlNullableID(sshCAID, "sshCAId"),
	)

	response := query.CreateGateway{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: address}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadGateway(ctx context.Context, gatewayID string) (*model.Gateway, error) {
	opr := resourceGateway.read()

	if gatewayID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(gatewayID))
	response := query.ReadGateway{}

	if err := client.query(ctx, &response, variables, opr, attr{id: gatewayID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateGateway(ctx context.Context, gateway *model.Gateway) (*model.Gateway, error) {
	opr := resourceGateway.update()

	if gateway.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if gateway.Address == "" {
		return nil, opr.apiError(ErrGraphqlAddressIsEmpty)
	}

	variables := newVars(
		gqlID(gateway.ID),
		gqlVar(gateway.Address, "address"),
		gqlID(gateway.RemoteNetworkID, "remoteNetworkId"),
		gqlID(gateway.X509CAID, "x509CAId"),
		gqlNullableID(gateway.SSHCAID, "sshCAId"),
	)

	response := query.UpdateGateway{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: gateway.ID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteGateway(ctx context.Context, gatewayID string) error {
	opr := resourceGateway.delete()

	if gatewayID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteGateway{}

	return client.mutate(ctx, &response, newVars(gqlID(gatewayID)), opr, attr{id: gatewayID})
}
